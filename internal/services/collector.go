package services

import (
	"errors"
	"math"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/evilmagics/dataset_collector/internal/config"
	"github.com/evilmagics/dataset_collector/internal/utils"
	"github.com/panjf2000/ants/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
)

type Collector struct {
	fs              afero.Fs
	conf            *config.Config
	datasetConf     *config.Dataset
	increments      map[utils.Category]*utils.Increment
	pool            *ants.MultiPool
	progressTotal   int64
	currentProgress int64
}

type Item struct {
	SrcFilename string
	SrcPath     string
	DstFilename string
	DstPath     string
	Data        []byte
}

type DatasetItem struct {
	Id     int
	SrcDir string
	DstDir string
	Label  *Item
	Image  *Item
	Cat    utils.Category
}

func (i *DatasetItem) SetNewFilename(id int) {
	var (
		ext     = path.Ext(i.Image.SrcFilename)
		newName = string(i.Cat) + "_" + strconv.Itoa(id)
	)
	i.Id = id
	i.Image.DstFilename = utils.RealFilename(newName, ext)
	i.Image.DstPath = utils.ImagePath(i.DstDir, utils.RealFilename(newName, ext))

	i.Label.DstFilename = utils.RealFilename(newName, ".txt")
	i.Label.DstPath = utils.LabelPath(i.DstDir, utils.RealFilename(newName, ".txt"))
}

func CreateDatasetItem(src, dst, imageFilename string, cat utils.Category) *DatasetItem {
	var (
		oldName = utils.Filename(imageFilename)
	)
	dst = path.Join(dst, string(cat))
	ds := &DatasetItem{
		SrcDir: src,
		DstDir: dst,
		Cat:    cat,
		Image: &Item{
			SrcFilename: imageFilename,
			SrcPath:     utils.ImagePath(src, imageFilename),
		},
		Label: &Item{
			SrcFilename: utils.RealFilename(oldName, ".txt"),
			SrcPath:     utils.LabelPath(src, utils.RealFilename(oldName, ".txt")),
		},
	}

	return ds
}

// NewCollector creates and initializes a new Collector with default filesystem, configuration,
// and increments for test, train, and validation dataset categories.
// It takes configuration and dataset configuration as parameters, with optional filesystem.
// Returns a pointer to the newly created Collector.
func NewCollector(conf *config.Config, datasetConf *config.Dataset, fs ...afero.Fs) (*Collector, error) {
	workers := int(math.Round(float64(conf.Workers) / 5))
	pool, err := ants.NewMultiPool(5, workers, ants.RoundRobin, ants.WithPreAlloc(true))
	if err != nil {
		return nil, err
	}

	return &Collector{
		fs:          afero.NewOsFs(),
		conf:        conf,
		datasetConf: datasetConf,
		pool:        pool,
		increments: map[utils.Category]*utils.Increment{
			utils.CategoryTest:  utils.NewIncrement(),
			utils.CategoryTrain: utils.NewIncrement(),
			utils.CategoryValid: utils.NewIncrement(),
		},
	}, nil
}

// GetIncrement returns and increments the counter for a specific dataset category.
// It increases the internal increment value for the given category and returns the new value.
func (c *Collector) GetIncrement(cat utils.Category) int { return c.increments[cat].Increase() }

// Collect orchestrates the dataset collection process by creating the destination folder,
// generating the dataset configuration, and collecting data from each configured source.
// It handles creating the destination directory, generating the data configuration file,
// and processing each source by loading its configuration and collecting images.
// Returns an error if any step in the collection process fails.
func (c Collector) CollectAll() (err error) {
	// Create destination folder
	if err = c.CreateDestFolder(); err != nil {
		log.Fatal().Err(err).Msg("Failed create destination folder")
		return err
	}

	// Create data.yaml on root destination folder
	if err := c.CreateConfig(); err != nil {
		log.Fatal().Err(err).Msg("Failed create dataset config!")
		return err
	}

	for _, s := range c.conf.Sources {
		if err := s.LoadDatasetConfig(c.fs); err != nil {
			log.Fatal().Err(err).Str("Src", s.Src).Msg("Dataset config can't loaded")
		}

		if err := c.Collect(s); err != nil {
			log.Fatal().Err(err).Str("Src", s.Src).Msg("Failed collect from source")
		}
	}

	return nil
}

// CreateConfig generates a new dataset configuration using the collector's configured classes
// and saves it to the destination directory using the collector's filesystem.
func (c Collector) CreateConfig() error {
	conf := config.NewDataset(c.conf.Classes...)

	return config.SaveDataset(c.fs, *conf, c.conf.Dest)
}

// Collect collecting images and label folder and renaming to destination
func (c Collector) Collect(src config.Source) error {
	entry, err := afero.ReadDir(c.fs, src.Src)
	if err != nil {
		return err
	}

	// Lookup for each folder on root source directory
	// Ensures the datasets folder on cross-named folder category
	for _, e := range entry {
		if !e.IsDir() {
			continue
		}

		dir := path.Join(src.Src, e.Name())

		cat := utils.FindCategory(e.Name())
		if cat == nil {
			continue
		}

		log.Info().Any("Name", e.Name()).Str("Path", dir).Msg("Collecting dataset on folder.")
		if err := c.collectDataset(src, *cat, dir); err != nil {
			log.Warn().Err(err).Str("Dir Path", dir).Msg("Failed collect from folder")
		}

	}

	return nil
}

func (c Collector) collectItem(item *DatasetItem, src config.Source) (err error) {
	// Read label file
	item.Label.Data, err = afero.ReadFile(c.fs, item.Label.SrcPath)
	if err != nil {
		return err
	}

	// Sync class index
	item.Label.Data, err = c.SyncClasses(item.Label.Data, src)
	if err != nil {
		return err
	}

	// Read image file
	item.Image.Data, err = afero.ReadFile(c.fs, item.Image.SrcPath)
	if err != nil {
		return err
	}

	// Set destination objects
	item.SetNewFilename(c.GetIncrement(item.Cat))

	if err := c.writeToDest(item); err != nil {
		return err
	}

	return nil
}

func (c Collector) collectDataset(src config.Source, cat utils.Category, dir string) (err error) {
	defer func() {
		if err != nil {
			log.Warn().Err(err).Send()
		}
	}()

	// read image filename as key
	images, err := afero.ReadDir(c.fs, path.Join(dir, "images"))
	if err != nil {
		return err
	}

	for _, f := range images {
		c.pool.Submit(func() {
			item := CreateDatasetItem(dir, c.conf.Dest, f.Name(), cat)
			if err = c.collectItem(item, src); err != nil {
				log.Warn().Err(err).Str("Path", utils.RightWrap(item.Image.SrcPath, 100)).Msg("Collecting item.")
				return
			}

			log.Info().Int("id", item.Id).Str("filename", item.Image.DstFilename).Msg("Dataset collected.")
		})
	}

	return nil
}

func (c Collector) write(path string, data []byte) error {
	f, err := c.fs.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	return err
}

func (c Collector) writeToDest(item *DatasetItem) (err error) {
	// Delete file on error
	defer func() {
		if err != nil {
			c.fs.Remove(item.Image.DstPath)
			c.fs.Remove(item.Label.DstPath)
		}
	}()

	if err = c.write(item.Image.DstPath, item.Image.Data); err != nil {
		return err
	}
	if err = c.write(item.Label.DstPath, item.Label.Data); err != nil {
		return err
	}

	return nil
}

func (c Collector) syncClasses(object string, src config.Source) (res string, err error) {
	obj := strings.Split(string(object), " ")
	if len(obj) > 0 {
		id, err := strconv.Atoi(obj[0])
		if err != nil {
			return res, err
		}
		// Find crossing class name from origin class
		origin := src.DatasetConfig.GetClassName(id)
		cross := src.ClassSync.GetCrossName(origin)

		if cross == nil {
			return res, errors.New("Cross class not found")
		}
		crossID := c.datasetConf.GetClassId(*cross)

		// Change first data with class index dest
		obj[0] = strconv.Itoa(crossID)

		return strings.Join(obj, " "), nil
	}
	return res, nil
}

// SyncClasses change original classes index to dest class index
// Example: 'vehicles' (0) -> 'car' (0)
// Example: 'vehicles' (0) -> 'car' (1)
func (c Collector) SyncClasses(body []byte, src config.Source) ([]byte, error) {
	// Split each line, this indicate to object count
	objects := strings.Split(string(body), "\n")
	newObjects := []string{}

	for _, o := range objects {
		res, err := c.syncClasses(o, src)
		if err != nil {
			continue
		}
		newObjects = append(newObjects, res)
	}

	if len(newObjects) == 0 {
		return nil, errors.New("No cross-object found")
	}

	// return updated data
	return []byte(strings.Join(newObjects, "\n")), nil
}

// CreateDestFolder creates the destination directory structure for training, testing, and validation datasets
// with separate subdirectories for images and labels. Returns an error if directory creation fails.
func (c Collector) CreateDestFolder() error {
	var err error

	// Create dest root directory
	err = c.fs.MkdirAll(c.conf.Dest, os.ModePerm)

	// Create training dataset directory on root directory
	err = c.fs.MkdirAll(path.Join(c.conf.Dest, "train", "images"), os.ModePerm)
	err = c.fs.MkdirAll(path.Join(c.conf.Dest, "train", "labels"), os.ModePerm)

	// Create testing dataset directory on root directory
	err = c.fs.MkdirAll(path.Join(c.conf.Dest, "test", "images"), os.ModePerm)
	err = c.fs.MkdirAll(path.Join(c.conf.Dest, "test", "labels"), os.ModePerm)

	// Create validation dataset directory on root directory
	err = c.fs.MkdirAll(path.Join(c.conf.Dest, "valid", "images"), os.ModePerm)
	err = c.fs.MkdirAll(path.Join(c.conf.Dest, "valid", "labels"), os.ModePerm)

	return err
}
