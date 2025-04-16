package utils

type Category string

const (
	CategoryTest  = "test"
	CategoryTrain = "train"
	CategoryValid = "valid"
)

var (
	categoryCrossName = map[string]Category{
		"test":       CategoryTest,
		"tests":      CategoryTest,
		"testing":    CategoryTest,
		"testings":   CategoryTest,
		"train":      CategoryTrain,
		"training":   CategoryTrain,
		"valid":      CategoryValid,
		"validation": CategoryValid,
	}
)

func FindCategory(name string) *Category {
	if c := categoryCrossName[name]; c != "" {
		return &c
	}
	return nil
}

func IsCategoryDetected(name string) bool {
	return FindCategory(name) != nil
}
