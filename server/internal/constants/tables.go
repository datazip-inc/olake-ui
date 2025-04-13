package constants

// Golang Enums
type TableType int64

const (
	UserTable TableType = iota
	SourceTable
	DestinationTable
	JobTable
	CatalogTable
)
