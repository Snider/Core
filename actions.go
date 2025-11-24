package core

// ActionServiceStartup is a message sent when the application's services are starting up.
// This provides a hook for services to perform initialization tasks.
type ActionServiceStartup struct{}

// ActionServiceShutdown is a message sent when the application is shutting down.
// This allows services to perform cleanup tasks, such as saving state or closing resources.
type ActionServiceShutdown struct{}
