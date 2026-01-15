package display

import (
	"fmt"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"
)

// NotificationOptions contains options for showing a notification.
type NotificationOptions struct {
	ID       string `json:"id,omitempty"`
	Title    string `json:"title"`
	Message  string `json:"message"`
	Subtitle string `json:"subtitle,omitempty"`
}

// SetNotifier sets the notifications service for native notifications.
func (s *Service) SetNotifier(notifier *notifications.NotificationService) {
	s.notifier = notifier
}

// ShowNotification displays a native system notification.
// Falls back to dialog if notifier is not available.
func (s *Service) ShowNotification(opts NotificationOptions) error {
	// Try native notification first
	if s.notifier != nil {
		// Generate ID if not provided
		id := opts.ID
		if id == "" {
			id = fmt.Sprintf("core-%d", time.Now().UnixNano())
		}

		return s.notifier.SendNotification(notifications.NotificationOptions{
			ID:       id,
			Title:    opts.Title,
			Subtitle: opts.Subtitle,
			Body:     opts.Message,
		})
	}

	// Fall back to dialog-based notification
	return s.showDialogNotification(opts)
}

// showDialogNotification shows a notification using dialogs as fallback.
func (s *Service) showDialogNotification(opts NotificationOptions) error {
	app := application.Get()
	if app == nil {
		return fmt.Errorf("application not available")
	}

	// Build message with optional subtitle
	msg := opts.Message
	if opts.Subtitle != "" {
		msg = opts.Subtitle + "\n\n" + msg
	}

	dialog := app.Dialog.Info()
	dialog.SetTitle(opts.Title)
	dialog.SetMessage(msg)
	dialog.Show()

	return nil
}

// ShowInfoNotification shows an info notification with a simple message.
func (s *Service) ShowInfoNotification(title, message string) error {
	return s.ShowNotification(NotificationOptions{
		Title:   title,
		Message: message,
	})
}

// ShowWarningNotification shows a warning notification.
func (s *Service) ShowWarningNotification(title, message string) error {
	app := application.Get()
	if app == nil {
		return fmt.Errorf("application not available")
	}

	dialog := app.Dialog.Warning()
	dialog.SetTitle(title)
	dialog.SetMessage(message)
	dialog.Show()

	return nil
}

// ShowErrorNotification shows an error notification.
func (s *Service) ShowErrorNotification(title, message string) error {
	app := application.Get()
	if app == nil {
		return fmt.Errorf("application not available")
	}

	dialog := app.Dialog.Error()
	dialog.SetTitle(title)
	dialog.SetMessage(message)
	dialog.Show()

	return nil
}

// RequestNotificationPermission requests permission for native notifications.
func (s *Service) RequestNotificationPermission() (bool, error) {
	if s.notifier == nil {
		return false, fmt.Errorf("notification service not available")
	}

	granted, err := s.notifier.RequestNotificationAuthorization()
	if err != nil {
		return false, fmt.Errorf("failed to request notification permission: %w", err)
	}

	return granted, nil
}

// CheckNotificationPermission checks if notifications are authorized.
func (s *Service) CheckNotificationPermission() (bool, error) {
	if s.notifier == nil {
		return false, fmt.Errorf("notification service not available")
	}

	return s.notifier.CheckNotificationAuthorization()
}
