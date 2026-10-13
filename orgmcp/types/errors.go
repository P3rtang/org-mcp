package orgmcp_types

const (
	NO_CHILDREN         = "ERROR | [%T]: with UID `%s` cannot have children."
	DUPLICATE_UID       = "ERROR | [%T]: with UID `%s` already contains a child with UID `%s`. Either update an existent table or choose a new `Name` to insert."
	COLUMN_OUT_OF_RANGE = "ERROR | [%T]: with UID `%s`, the column specified is out of range, column length: %d, got: %d"
)
