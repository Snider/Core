// Package bugseti provides services for the BugSETI distributed bug fixing application.
package bugseti

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"time"
)

// NotifyService handles desktop notifications.
type NotifyService struct {
	enabled bool
	sound   bool
}

// NewNotifyService creates a new NotifyService.
func NewNotifyService() *NotifyService {
	return &NotifyService{
		enabled: true,
		sound:   true,
	}
}

// ServiceName returns the service name for Wails.
func (n *NotifyService) ServiceName() string {
	return "NotifyService"
}

// SetEnabled enables or disables notifications.
func (n *NotifyService) SetEnabled(enabled bool) {
	n.enabled = enabled
}

// SetSound enables or disables notification sounds.
func (n *NotifyService) SetSound(sound bool) {
	n.sound = sound
}

// Notify sends a desktop notification.
func (n *NotifyService) Notify(title, message string) error {
	if !n.enabled {
		return nil
	}

	log.Printf("Notification: %s - %s", title, message)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var err error
	switch runtime.GOOS {
	case "darwin":
		err = n.notifyMacOS(ctx, title, message)
	case "linux":
		err = n.notifyLinux(ctx, title, message)
	case "windows":
		err = n.notifyWindows(ctx, title, message)
	default:
		err = fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if err != nil {
		log.Printf("Notification error: %v", err)
	}
	return err
}

// NotifyIssue sends a notification about a new issue.
func (n *NotifyService) NotifyIssue(issue *Issue) error {
	title := "New Issue Available"
	message := fmt.Sprintf("%s: %s", issue.Repo, issue.Title)
	return n.Notify(title, message)
}

// NotifyPRStatus sends a notification about a PR status change.
func (n *NotifyService) NotifyPRStatus(repo string, prNumber int, status string) error {
	title := "PR Status Update"
	message := fmt.Sprintf("%s #%d: %s", repo, prNumber, status)
	return n.Notify(title, message)
}

// notifyMacOS sends a notification on macOS using osascript.
func (n *NotifyService) notifyMacOS(ctx context.Context, title, message string) error {
	script := fmt.Sprintf(`display notification "%s" with title "%s"`, message, title)
	if n.sound {
		script += ` sound name "Glass"`
	}
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	return cmd.Run()
}

// notifyLinux sends a notification on Linux using notify-send.
func (n *NotifyService) notifyLinux(ctx context.Context, title, message string) error {
	args := []string{
		"--app-name=BugSETI",
		"--urgency=normal",
		title,
		message,
	}
	cmd := exec.CommandContext(ctx, "notify-send", args...)
	return cmd.Run()
}

// notifyWindows sends a notification on Windows using PowerShell.
func (n *NotifyService) notifyWindows(ctx context.Context, title, message string) error {
	script := fmt.Sprintf(`
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null

$template = @"
<toast>
    <visual>
        <binding template="ToastText02">
            <text id="1">%s</text>
            <text id="2">%s</text>
        </binding>
    </visual>
</toast>
"@

$xml = New-Object Windows.Data.Xml.Dom.XmlDocument
$xml.LoadXml($template)
$toast = [Windows.UI.Notifications.ToastNotification]::new($xml)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("BugSETI").Show($toast)
`, title, message)

	cmd := exec.CommandContext(ctx, "powershell", "-Command", script)
	return cmd.Run()
}

// NotifyWithAction sends a notification with an action button (platform-specific).
func (n *NotifyService) NotifyWithAction(title, message, actionLabel string) error {
	if !n.enabled {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch runtime.GOOS {
	case "darwin":
		// macOS: Use terminal-notifier if available for actions
		if _, err := exec.LookPath("terminal-notifier"); err == nil {
			cmd := exec.CommandContext(ctx, "terminal-notifier",
				"-title", title,
				"-message", message,
				"-appIcon", "NSApplication",
				"-actions", actionLabel,
				"-group", "BugSETI")
			return cmd.Run()
		}
		return n.notifyMacOS(ctx, title, message)

	case "linux":
		// Linux: Use notify-send with action
		args := []string{
			"--app-name=BugSETI",
			"--urgency=normal",
			"--action=open=" + actionLabel,
			title,
			message,
		}
		cmd := exec.CommandContext(ctx, "notify-send", args...)
		return cmd.Run()

	default:
		return n.Notify(title, message)
	}
}

// NotifyProgress sends a notification with a progress indicator.
func (n *NotifyService) NotifyProgress(title, message string, progress int) error {
	if !n.enabled {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch runtime.GOOS {
	case "linux":
		// Linux supports progress hints
		args := []string{
			"--app-name=BugSETI",
			"--hint=int:value:" + fmt.Sprintf("%d", progress),
			title,
			message,
		}
		cmd := exec.CommandContext(ctx, "notify-send", args...)
		return cmd.Run()

	default:
		// Other platforms: include progress in message
		messageWithProgress := fmt.Sprintf("%s (%d%%)", message, progress)
		return n.Notify(title, messageWithProgress)
	}
}

// PlaySound plays a notification sound.
func (n *NotifyService) PlaySound() error {
	if !n.sound {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	switch runtime.GOOS {
	case "darwin":
		cmd := exec.CommandContext(ctx, "afplay", "/System/Library/Sounds/Glass.aiff")
		return cmd.Run()

	case "linux":
		// Try paplay (PulseAudio), then aplay (ALSA)
		if _, err := exec.LookPath("paplay"); err == nil {
			cmd := exec.CommandContext(ctx, "paplay", "/usr/share/sounds/freedesktop/stereo/complete.oga")
			return cmd.Run()
		}
		if _, err := exec.LookPath("aplay"); err == nil {
			cmd := exec.CommandContext(ctx, "aplay", "-q", "/usr/share/sounds/alsa/Front_Center.wav")
			return cmd.Run()
		}
		return nil

	case "windows":
		script := `[console]::beep(800, 200)`
		cmd := exec.CommandContext(ctx, "powershell", "-Command", script)
		return cmd.Run()

	default:
		return nil
	}
}
