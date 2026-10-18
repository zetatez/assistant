package psl

import (
	"fmt"
	"net"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"

	"assistant/internal/news"
	"assistant/internal/platform"
	"assistant/pkg/llm"
	"assistant/pkg/xlog"

	"github.com/spf13/viper"
)

var (
	onceConfig sync.Once
	config     *Config
)

func GetConfig() *Config { return config }

func InitConfig() error {
	var err error
	onceConfig.Do(func() {
		config, err = loadConfig()
	})
	return err
}

func loadConfig() (*Config, error) {
	home, _ := os.UserHomeDir()
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.SetEnvPrefix("APP")
	v.AutomaticEnv()
	v.AddConfigPath(".")
	v.AddConfigPath(path.Join(home, ".config/assistant"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read in config: %v", err)
	}
	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("failed unmarshal config: %v", err)
	}
	c.resolveEnv()
	c.applyDefaults()
	platform.SetOverrides(c.Settings.Commands)
	if err := c.expandPaths(); err != nil {
		return nil, fmt.Errorf("expand paths: %w", err)
	}
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	return &c, nil
}

type Config struct {
	App        AppConfig        `mapstructure:"app"`
	Log        xlog.LogConfig   `mapstructure:"log"`
	LLM        llm.Config       `mapstructure:"llm"`
	Settings   SettingsConfig   `mapstructure:"settings"`
	Background BackgroundConfig `mapstructure:"background"`
	News       news.Config      `mapstructure:"news"`
}

type AppConfig struct {
	Name         string `mapstructure:"name"`
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Interface    string `mapstructure:"interface"`
	RootUsername string `mapstructure:"root_username"`
	RootPassword string `mapstructure:"root_password"`
}

type SettingsConfig struct {
	DefaultMonitor         string `mapstructure:"default_monitor"`
	DirSnip                string `mapstructure:"dir_snip"`
	DirWallpaper           string `mapstructure:"dir_wallpaper"`
	DirWorkingLogbook      string `mapstructure:"dir_working_logbook"`
	DirAIOutput            string `mapstructure:"dir_ai_output"`
	PathKeyboardBrightness string `mapstructure:"filepath_keyboard_brightness"`
	PathSSHSecret          string `mapstructure:"filepath_ssh_secret"`
	DefaultTerminal        string `mapstructure:"default_terminal"`
	VPN                    string `mapstructure:"vpn"`
	CaptureScreenDevice    string `mapstructure:"capture_screen_device"`
	CaptureAudioDevice     string `mapstructure:"capture_audio_device"`
	CaptureVideoDevice     string `mapstructure:"capture_video_device"`
	// Commands overrides built-in platform commands per intent, e.g.
	//   commands:
	//     keyboard_light_toggle: "kbdlight toggle"
	Commands        map[string]string `mapstructure:"commands"`
	FeishuAppID     string            `mapstructure:"feishu_app_id"`
	FeishuAppSecret string            `mapstructure:"feishu_app_secret"`
	FeishuChatID    string            `mapstructure:"feishu_chat_id"`
}

type BackgroundConfig struct {
	Enabled bool             `mapstructure:"enabled"`
	Procs   []BackgroundProc `mapstructure:"procs"`
}

type BackgroundProc struct {
	Name      string `mapstructure:"name"`
	Command   string `mapstructure:"command"`
	Precursor string `mapstructure:"precursor"`
}

func (c *Config) applyDefaults() {
	if c.App.Host == "" {
		c.App.Host = "0.0.0.0"
	}
	if c.App.Port == 0 {
		c.App.Port = 4321
	}
	if c.App.Interface == "" {
		c.App.Interface = platform.Get().DefaultInterfaceName()
	}
	if c.LLM.Timeout == 0 {
		c.LLM.Timeout = 300
	}
	if c.Settings.DefaultMonitor == "" {
		c.Settings.DefaultMonitor = defaultMonitor()
	}
	if c.Settings.DirWallpaper == "" {
		c.Settings.DirWallpaper = "~/Pictures/wallpapers"
	}
	if c.Settings.DirWorkingLogbook == "" {
		c.Settings.DirWorkingLogbook = "~/git/working/logbook"
	}
	if c.Settings.DirAIOutput == "" {
		c.Settings.DirAIOutput = "~/git/test"
	}
	if c.Settings.PathKeyboardBrightness == "" {
		c.Settings.PathKeyboardBrightness = defaultKeyboardBrightnessPath()
	}
	if c.Settings.DefaultTerminal == "" {
		c.Settings.DefaultTerminal = defaultTerminal()
	}
	if c.Settings.DirSnip == "" {
		c.Settings.DirSnip = "~/git/obsidian/.snippets"
	}
	if len(c.Background.Procs) == 0 {
		c.Background.Procs = defaultBackgroundProcs()
	}
}

// defaultBackgroundProcs returns the daemons worth supervising on the current
// OS. On macOS the dwm/X11 daemons do not exist, so nothing is started unless
// the user configures it explicitly.
func defaultBackgroundProcs() []BackgroundProc {
	if platform.Get().GOOS() != platform.Linux {
		return nil
	}
	home, _ := os.UserHomeDir()
	return []BackgroundProc{
		{Name: "dwmblocks", Command: "dwmblocks"},
		{Name: "picom", Command: "picom --config " + home + "/.config/picom/picom.conf"},
		{Name: "dunst", Command: "dunst"},
	}
}

func defaultMonitor() string {
	if platform.Get().GOOS() == platform.Darwin {
		return "Built-in"
	}
	return "eDP-1"
}

func defaultTerminal() string {
	if platform.Get().GOOS() == platform.Darwin {
		return "Terminal"
	}
	return "st"
}

func defaultKeyboardBrightnessPath() string {
	if platform.Get().GOOS() == platform.Darwin {
		return ""
	}
	return "/sys/class/leds/tpacpi::kbd_backlight/brightness"
}

func (c *Config) expandPaths() error {
	home, _ := os.UserHomeDir()
	for _, p := range []*string{
		&c.Log.Filename,
		&c.Settings.DirWallpaper,
		&c.Settings.DirWorkingLogbook,
		&c.Settings.DirAIOutput,
		&c.Settings.PathSSHSecret,
		&c.Settings.PathKeyboardBrightness,
		&c.Settings.DirSnip,
	} {
		*p = expandHomePath(*p, home)
	}
	return nil
}

func expandHomePath(p, home string) string {
	if p == "" {
		return p
	}
	if p == "~" {
		return home
	}
	if strings.HasPrefix(p, "~/") {
		return path.Join(home, p[2:])
	}
	return p
}

func (c *Config) Validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}
	if c.App.Host != "" && net.ParseIP(c.App.Host) == nil && c.App.Host != "0.0.0.0" {
		return fmt.Errorf("app.host must be a valid IP address: %s", c.App.Host)
	}
	return nil
}

func (c *Config) resolveEnv() {
	envPH := regexp.MustCompile(`\$\{(\w+)\}`)
	expand := func(p *string) { *p = expandEnvPH(*p, envPH) }
	expand(&c.Settings.FeishuAppID)
	expand(&c.Settings.FeishuAppSecret)
	expand(&c.Settings.FeishuChatID)
	expand(&c.LLM.APIKey)
}

func expandEnvPH(s string, envPH *regexp.Regexp) string {
	if s == "" {
		return s
	}
	return envPH.ReplaceAllStringFunc(s, func(m string) string {
		groups := envPH.FindStringSubmatch(m)
		if len(groups) < 2 {
			return m
		}
		return os.Getenv(groups[1])
	})
}
