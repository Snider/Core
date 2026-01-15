package electroncompat

// DBService provides key-value storage operations.
// This corresponds to the DB IPC service from the Electron app.
type DBService struct{}

// NewDBService creates a new DBService instance.
func NewDBService() *DBService {
	return &DBService{}
}

// Open opens a database.
func (s *DBService) Open(name string) error {
	return notImplemented("DB", "open")
}

// Close closes the database.
func (s *DBService) Close() error {
	return notImplemented("DB", "close")
}

// Put stores a value by key.
func (s *DBService) Put(key string, value any) error {
	return notImplemented("DB", "put")
}

// Get retrieves a value by key.
func (s *DBService) Get(key string) (any, error) {
	return nil, notImplemented("DB", "get")
}

// Del deletes a value by key.
func (s *DBService) Del(key string) error {
	return notImplemented("DB", "del")
}

// GetUserDir returns the user data directory.
func (s *DBService) GetUserDir() (string, error) {
	return "", notImplemented("DB", "getUserDir")
}
