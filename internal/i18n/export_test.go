package i18n

// CatalogKeys exposes catalogKeys to the external test package, which checks
// the catalog and the source for unused and missing keys.
func CatalogKeys(language string) []string { return catalogKeys(language) }
