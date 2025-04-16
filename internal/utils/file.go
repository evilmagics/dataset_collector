package utils

import (
	"path"
	"strings"
)

func ChangeFileExt(src, ext string) string {
	return strings.TrimSuffix(src, path.Ext(src)) + ext
}

// Filename return filename (without extension) from path
func Filename(src string) string {
	return strings.TrimSuffix(src, path.Ext(src))
}

// RealFilename return combination with filename and extension
func RealFilename(str, ext string) string {
	return str + ext
}

func LabelPath(dir, filename string, cat ...Category) string {
	if len(cat) > 0 {
		return path.Join(dir, string(cat[0]), "labels", filename)
	}
	return path.Join(dir, "labels", filename)
}

func ImagePath(dir, filename string, cat ...Category) string {
	if len(cat) > 0 {
		return path.Join(dir, string(cat[0]), "images", filename)
	}
	return path.Join(dir, "images", filename)
}
