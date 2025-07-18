// config.go: This file contains the configuration for the BirdNET-Go application. It defines the settings struct and functions to load and save the settings.
package conf

import (
	"crypto/rand"
	"embed"
	"encoding/base64"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
	"github.com/tphakala/birdnet-go/internal/errors"
	"gopkg.in/yaml.v3"
)

//go:embed config.yaml
var configFiles embed.FS

// EqualizerFilter is a struct for equalizer filter settings
type EqualizerFilter struct {
	Type      string // e.g., "LowPass", "HighPass", "BandPass", etc.
	Frequency float64
	Q         float64
	Gain      float64 // Only used for certain filter types like Peaking
	Width     float64 // Only used for certain filter types like BandPass and BandReject
	Passes    int     // Filter passes for added attenuation or gain
}

// EqualizerSettings is a struct for audio EQ settings
type EqualizerSettings struct {
	Enabled bool              // global flag to enable/disable equalizer filters
	Filters []EqualizerFilter // equalizer filter configuration
}

type ExportSettings struct {
	Debug     bool              // true to enable audio export debug
	Enabled   bool              // export audio clips containing indentified bird calls
	Path      string            // path to audio clip export directory
	Type      string            // audio file type, wav, mp3 or flac
	Bitrate   string            // bitrate for audio export
	Retention RetentionSettings // retention settings
}

type RetentionSettings struct {
	Debug            bool   // true to enable retention debug
	Policy           string // retention policy, "none", "age" or "usage"
	MaxAge           string // maximum age of audio clips to keep
	MaxUsage         string // maximum disk usage percentage before cleanup
	MinClips         int    // minimum number of clips per species to keep
	KeepSpectrograms bool   // true to keep spectrograms
}

// AudioSettings contains settings for audio processing and export.
// SoundLevelSettings contains settings for sound level monitoring
type SoundLevelSettings struct {
	Enabled              bool `yaml:"enabled" mapstructure:"enabled"`                               // true to enable sound level monitoring
	Interval             int  `yaml:"interval" mapstructure:"interval"`                             // measurement interval in seconds (default: 10)
	Debug                bool `yaml:"debug" mapstructure:"debug"`                                   // true to enable debug logging for sound level monitoring
	DebugRealtimeLogging bool `yaml:"debug_realtime_logging" mapstructure:"debug_realtime_logging"` // true to log debug messages for every realtime update, false to log only at configured interval
}

type AudioSettings struct {
	Source          string             // audio source to use for analysis
	FfmpegPath      string             // path to ffmpeg, runtime value
	SoxPath         string             // path to sox, runtime value
	SoxAudioTypes   []string           `yaml:"-"` // supported audio types of sox, runtime value
	StreamTransport string             // preferred transport for audio streaming: "auto", "sse", or "ws"
	Export          ExportSettings     // export settings
	SoundLevel      SoundLevelSettings // sound level monitoring settings
	UseAudioCore    bool               // true to use new audiocore package instead of myaudio

	Equalizer EqualizerSettings // equalizer settings
}
type Thumbnails struct {
	Debug          bool   // true to enable debug mode
	Summary        bool   // show thumbnails on summary table
	Recent         bool   // show thumbnails on recent table
	ImageProvider  string // preferred image provider: "auto", "wikimedia", "avicommons"
	FallbackPolicy string // fallback policy: "none", "all" - try all available providers if preferred fails
}

// Dashboard contains settings for the web dashboard.
type Dashboard struct {
	Thumbnails   Thumbnails // thumbnails settings
	SummaryLimit int        // limit for the number of species shown in the summary table
}

// DynamicThresholdSettings contains settings for dynamic threshold adjustment.
type DynamicThresholdSettings struct {
	Enabled    bool    // true to enable dynamic threshold
	Debug      bool    // true to enable debug mode
	Trigger    float64 // trigger threshold for dynamic threshold
	Min        float64 // minimum threshold for dynamic threshold
	ValidHours int     // number of hours to consider for dynamic threshold
}

// RetrySettings contains common settings for retry mechanisms
type RetrySettings struct {
	Enabled           bool    // true to enable retry mechanism
	MaxRetries        int     // maximum number of retry attempts
	InitialDelay      int     // initial delay before first retry in seconds
	MaxDelay          int     // maximum delay between retries in seconds
	BackoffMultiplier float64 // multiplier for exponential backoff
}

// BirdweatherSettings contains settings for BirdWeather API integration.
type BirdweatherSettings struct {
	Enabled          bool          // true to enable birdweather uploads
	Debug            bool          // true to enable debug mode
	ID               string        // birdweather ID
	Threshold        float64       // threshold for prediction confidence for uploads
	LocationAccuracy float64       // accuracy of location in meters
	RetrySettings    RetrySettings // settings for retry mechanism
}

// WeatherSettings contains all weather-related settings
type WeatherSettings struct {
	Provider     string              // "none", "yrno" or "openweather"
	PollInterval int                 // weather data polling interval in minutes
	Debug        bool                // true to enable debug mode
	OpenWeather  OpenWeatherSettings // OpenWeather integration settings
}

// OpenWeatherSettings contains settings for OpenWeather integration.
type OpenWeatherSettings struct {
	Enabled  bool   // true to enable OpenWeather integration, for legacy support
	APIKey   string // OpenWeather API key
	Endpoint string // OpenWeather API endpoint
	Units    string // units of measurement: standard, metric, or imperial
	Language string // language code for the response
}

// PrivacyFilterSettings contains settings for the privacy filter.
type PrivacyFilterSettings struct {
	Debug      bool    // true to enable debug mode
	Enabled    bool    // true to enable privacy filter
	Confidence float32 // confidence threshold for human detection
}

// DogBarkFilterSettings contains settings for the dog bark filter.
type DogBarkFilterSettings struct {
	Debug      bool     // true to enable debug mode
	Enabled    bool     // true to enable dog bark filter
	Confidence float32  // confidence threshold for dog bark detection
	Remember   int      // how long we should remember bark for filtering?
	Species    []string // species list for filtering
}

// RTSPHealthSettings contains settings for RTSP stream health monitoring.
type RTSPHealthSettings struct {
	HealthyDataThreshold int // seconds before stream considered unhealthy (default: 60)
	MonitoringInterval   int // health check interval in seconds (default: 30)
}

// RTSPSettings contains settings for RTSP streaming.
type RTSPSettings struct {
	Transport        string              // RTSP Transport Protocol
	URLs             []string            // RTSP stream URL
	Health           RTSPHealthSettings  // health monitoring settings
	FFmpegParameters []string            // optional custom FFmpeg parameters
}

// MQTTSettings contains settings for MQTT integration.
type MQTTSettings struct {
	Enabled       bool            // true to enable MQTT
	Debug         bool            // true to enable MQTT debug
	Broker        string          // MQTT broker URL
	Topic         string          // MQTT topic
	Username      string          // MQTT username
	Password      string          // MQTT password
	Retain        bool            // true to retain messages
	RetrySettings RetrySettings   // settings for retry mechanism
	TLS           MQTTTLSSettings // TLS/SSL configuration
}

// MQTTTLSSettings contains TLS/SSL configuration for secure MQTT connections
type MQTTTLSSettings struct {
	Enabled            bool   // true to enable TLS (auto-detected from broker URL)
	InsecureSkipVerify bool   // true to skip certificate verification (for self-signed certs)
	CACert             string `yaml:"cacert,omitempty"`     // path to CA certificate file (managed internally)
	ClientCert         string `yaml:"clientcert,omitempty"` // path to client certificate file (managed internally)
	ClientKey          string `yaml:"clientkey,omitempty"`  // path to client key file (managed internally)
}

// TelemetrySettings contains settings for telemetry.
type TelemetrySettings struct {
	Enabled bool   // true to enable Prometheus compatible telemetry endpoint
	Listen  string // IP address and port to listen on
}

// MonitoringSettings contains settings for system resource monitoring
type MonitoringSettings struct {
	Enabled              bool                  // true to enable system resource monitoring
	CheckInterval        int                   // interval in seconds between resource checks
	CriticalResendInterval int                 // interval in minutes between critical alert resends (default: 30)
	HysteresisPercent    float64               // hysteresis percentage for state transitions (default: 5.0)
	CPU                  ThresholdSettings     // CPU usage thresholds
	Memory               ThresholdSettings     // Memory usage thresholds
	Disk                 DiskThresholdSettings // Disk usage thresholds
}

// ThresholdSettings contains warning and critical thresholds
type ThresholdSettings struct {
	Enabled  bool    // true to enable monitoring for this resource
	Warning  float64 // warning threshold percentage
	Critical float64 // critical threshold percentage
}

// DiskThresholdSettings contains disk monitoring configuration for multiple paths
type DiskThresholdSettings struct {
	Enabled  bool     // true to enable disk monitoring
	Warning  float64  // warning threshold percentage
	Critical float64  // critical threshold percentage
	Paths    []string // filesystem paths to monitor
}

// SentrySettings contains settings for Sentry error tracking
type SentrySettings struct {
	Enabled bool // true to enable Sentry error tracking (opt-in)
	Debug   bool // true to enable transparent telemetry logging
}

// RealtimeSettings contains all settings related to realtime processing.
type RealtimeSettings struct {
	Interval         int                      // minimum interval between log messages in seconds
	ProcessingTime   bool                     // true to report processing time for each prediction
	Audio            AudioSettings            // Audio processing settings
	Dashboard        Dashboard                // Dashboard settings
	DynamicThreshold DynamicThresholdSettings // Dynamic threshold settings
	Log              struct {
		Enabled bool   // true to enable OBS chat log
		Path    string // path to OBS chat log
	}
	Birdweather   BirdweatherSettings   // Birdweather integration settings
	OpenWeather   OpenWeatherSettings   `yaml:"-"` // OpenWeather integration settings
	PrivacyFilter PrivacyFilterSettings // Privacy filter settings
	DogBarkFilter DogBarkFilterSettings // Dog bark filter settings
	RTSP          RTSPSettings          // RTSP settings
	MQTT          MQTTSettings          // MQTT settings
	Telemetry     TelemetrySettings     // Telemetry settings
	Monitoring    MonitoringSettings    // System resource monitoring settings
	Species       SpeciesSettings       // Custom thresholds and actions for species
	Weather       WeatherSettings       // Weather provider related settings
}

// SpeciesAction represents a single action configuration
type SpeciesAction struct {
	Type            string   `yaml:"type"`            // Type of action (ExecuteCommand, etc)
	Command         string   `yaml:"command"`         // Path to the command to execute
	Parameters      []string `yaml:"parameters"`      // Action parameters
	ExecuteDefaults bool     `yaml:"executeDefaults"` // Whether to also execute default actions
}

// SpeciesConfig represents configuration for a specific species
type SpeciesConfig struct {
	Threshold float64         `yaml:"threshold"`          // Confidence threshold
	Interval  int             `yaml:"interval,omitempty"` // New field: Custom interval in seconds
	Actions   []SpeciesAction `yaml:"actions"`            // List of actions to execute
}

// RealtimeSpeciesSettings contains all species-specific settings
type SpeciesSettings struct {
	Include []string                 `yaml:"include"` // Always include these species
	Exclude []string                 `yaml:"exclude"` // Always exclude these species
	Config  map[string]SpeciesConfig `yaml:"config"`  // Per-species configuration
}

// ActionConfig holds configuration details for a specific action.
type ActionConfig struct {
	Type       string   // Type of the action (e.g. ExecuteScript which is only type for now)
	Parameters []string // List of parameters for the action
}

// InputConfig holds settings for file or directory analysis
type InputConfig struct {
	Path      string `yaml:"-"` // path to input file or directory
	Recursive bool   `yaml:"-"` // true for recursive directory analysis
	Watch     bool   `yaml:"-"` // true to watch directory for new files
}

type BirdNETConfig struct {
	Debug       bool                // true to enable debug mode
	Sensitivity float64             // birdnet analysis sigmoid sensitivity
	Threshold   float64             // threshold for prediction confidence to report
	Overlap     float64             // birdnet analysis overlap between chunks
	Longitude   float64             // longitude of recording location for prediction filtering
	Latitude    float64             // latitude of recording location for prediction filtering
	Threads     int                 // number of CPU threads to use for analysis
	Locale      string              // language to use for labels
	RangeFilter RangeFilterSettings // range filter settings
	ModelPath   string              // path to external model file (empty for embedded)
	LabelPath   string              // path to external label file (empty for embedded)
	Labels      []string            `yaml:"-"` // list of available species labels, runtime value
	UseXNNPACK  bool                // true to use XNNPACK delegate for inference acceleration
}

// RangeFilterSettings contains settings for the range filter
type RangeFilterSettings struct {
	Debug       bool      // true to enable debug mode
	Model       string    // range filter model model
	Threshold   float32   // rangefilter species occurrence threshold
	Species     []string  `yaml:"-"` // list of included species, runtime value
	LastUpdated time.Time `yaml:"-"` // last time the species list was updated, runtime value
}

// BasicAuth holds settings for the password authentication
type BasicAuth struct {
	Enabled        bool          // true to enable password authentication
	Password       string        // password for admin interface
	ClientID       string        // client id for OAuth2
	ClientSecret   string        // client secret for OAuth2
	RedirectURI    string        // redirect uri for OAuth2
	AuthCodeExp    time.Duration // duration for authorization code
	AccessTokenExp time.Duration // duration for access token
}

// SocialProvider holds settings for an OAuth2 identity provider
type SocialProvider struct {
	Enabled      bool   // true to enable social provider
	ClientID     string // client id for OAuth2
	ClientSecret string // client secret for OAuth2
	RedirectURI  string // redirect uri for OAuth2
	UserId       string // valid user id for OAuth2
}

type AllowSubnetBypass struct {
	Enabled bool   // true to enable subnet bypass
	Subnet  string // disable OAuth2 in subnet
}

// SecurityConfig handles all security-related settings and validations
// for the application, including authentication, TLS, and access control.
type Security struct {
	Debug bool // true to enable debug mode

	// Host is the primary hostname used for TLS certificates
	// and OAuth redirect URLs. Required when using AutoTLS or
	// authentication providers. Used to form the redirect URIs.
	Host string

	// AutoTLS enables automatic TLS certificate management using
	// Let's Encrypt. Requires Host to be set and port 80/443 access.
	AutoTLS bool

	RedirectToHTTPS   bool              // true to redirect to HTTPS
	AllowSubnetBypass AllowSubnetBypass // subnet bypass configuration
	BasicAuth         BasicAuth         // password authentication configuration
	GoogleAuth        SocialProvider    // Google OAuth2 configuration
	GithubAuth        SocialProvider    // Github OAuth2 configuration
	SessionSecret     string            // secret for session cookie
	SessionDuration   time.Duration     // duration for browser session cookies
}

type WebServerSettings struct {
	Debug      bool               // true to enable debug mode
	Enabled    bool               // true to enable web server
	Port       string             // port for web server
	Log        LogConfig          // logging configuration for web server
	LiveStream LiveStreamSettings // live stream configuration
}

type LiveStreamSettings struct {
	Debug          bool   // true to enable debug mode
	BitRate        int    // bitrate for live stream in kbps
	SampleRate     int    // sample rate for live stream in Hz
	SegmentLength  int    // length of each segment in seconds
	FfmpegLogLevel string // log level for ffmpeg
}

// BackupRetention defines backup retention policy
type BackupRetention struct {
	MaxAge     string `yaml:"maxage"`     // Duration string for the maximum age of backups to keep (e.g., "30d" for 30 days, "6m" for 6 months, "1y" for 1 year). Backups older than this may be deleted.
	MaxBackups int    `yaml:"maxbackups"` // Maximum total number of backups to keep for a given source. If 0, no limit by count (only by age or MinBackups).
	MinBackups int    `yaml:"minbackups"` // Minimum number of recent backups to keep for a given source, regardless of their age. This ensures a baseline number of backups are always available.
}

// BackupTargetSettings is an interface for type-safe backup target configuration
type BackupTargetSettings interface {
	Validate() error
}

// LocalBackupSettings defines settings for local filesystem backup target
type LocalBackupSettings struct {
	Path string `yaml:"path"` // Local filesystem path where backups will be stored
}

// Validate validates local backup settings
func (s *LocalBackupSettings) Validate() error {
	if s.Path == "" {
		return fmt.Errorf("local backup path cannot be empty")
	}
	return nil
}

// FTPBackupSettings defines settings for FTP backup target
type FTPBackupSettings struct {
	Host     string `yaml:"host"`     // FTP server hostname or IP address
	Port     int    `yaml:"port"`     // FTP server port (default: 21)
	Username string `yaml:"username"` // FTP username
	Password string `yaml:"password"` // FTP password
	Path     string `yaml:"path"`     // Remote path on FTP server
	UseTLS   bool   `yaml:"usetls"`   // Use FTPS (FTP over TLS)
}

// Validate validates FTP backup settings
func (s *FTPBackupSettings) Validate() error {
	if s.Host == "" {
		return fmt.Errorf("FTP host cannot be empty")
	}
	if s.Port == 0 {
		s.Port = 21 // Set default port
	}
	return nil
}

// SFTPBackupSettings defines settings for SFTP backup target
type SFTPBackupSettings struct {
	Host           string `yaml:"host"`           // SFTP server hostname or IP address
	Port           int    `yaml:"port"`           // SFTP server port (default: 22)
	Username       string `yaml:"username"`       // SFTP username
	Password       string `yaml:"password"`       // SFTP password (optional if using key)
	PrivateKeyPath string `yaml:"privatekeypath"` // Path to private key file (optional)
	Path           string `yaml:"path"`           // Remote path on SFTP server
}

// Validate validates SFTP backup settings
func (s *SFTPBackupSettings) Validate() error {
	if s.Host == "" {
		return fmt.Errorf("SFTP host cannot be empty")
	}
	if s.Port == 0 {
		s.Port = 22 // Set default port
	}
	if s.Username == "" {
		return fmt.Errorf("SFTP username cannot be empty")
	}
	return nil
}

// S3BackupSettings defines settings for S3-compatible backup target
type S3BackupSettings struct {
	Endpoint        string `yaml:"endpoint"`        // S3 endpoint URL
	Region          string `yaml:"region"`          // AWS region
	Bucket          string `yaml:"bucket"`          // S3 bucket name
	AccessKeyID     string `yaml:"accesskeyid"`     // AWS access key ID
	SecretAccessKey string `yaml:"secretaccesskey"` // AWS secret access key
	Prefix          string `yaml:"prefix"`          // Object key prefix
	UseSSL          bool   `yaml:"usessl"`          // Use SSL/TLS (default: true)
}

// Validate validates S3 backup settings
func (s *S3BackupSettings) Validate() error {
	if s.Bucket == "" {
		return fmt.Errorf("S3 bucket name cannot be empty")
	}
	if s.Region == "" {
		return fmt.Errorf("S3 region cannot be empty")
	}
	return nil
}

// RsyncBackupSettings defines settings for rsync backup target
type RsyncBackupSettings struct {
	Host       string   `yaml:"host"`       // Remote host (optional for local rsync)
	Port       int      `yaml:"port"`       // SSH port for remote rsync (default: 22)
	Username   string   `yaml:"username"`   // SSH username for remote rsync
	Path       string   `yaml:"path"`       // Destination path
	SSHKeyPath string   `yaml:"sshkeypath"` // Path to SSH private key
	Options    []string `yaml:"options"`    // Additional rsync options
}

// Validate validates rsync backup settings
func (s *RsyncBackupSettings) Validate() error {
	if s.Path == "" {
		return fmt.Errorf("rsync path cannot be empty")
	}
	if s.Host != "" && s.Port == 0 {
		s.Port = 22 // Set default SSH port for remote rsync
	}
	return nil
}

// GoogleDriveBackupSettings defines settings for Google Drive backup target
type GoogleDriveBackupSettings struct {
	CredentialsPath string `yaml:"credentialspath"` // Path to Google service account credentials JSON
	FolderID        string `yaml:"folderid"`        // Google Drive folder ID where backups will be stored
}

// Validate validates Google Drive backup settings
func (s *GoogleDriveBackupSettings) Validate() error {
	if s.CredentialsPath == "" {
		return fmt.Errorf("google drive credentials path cannot be empty")
	}
	return nil
}

// BackupTarget defines settings for a backup target
type BackupTarget struct {
	Type     string         `yaml:"type"`     // Specifies the type of the backup target (e.g., "local", "s3", "ftp", "sftp"). This determines the storage mechanism.
	Enabled  bool           `yaml:"enabled"`  // If true, this backup target will be used for storing backups. At least one target should be enabled for backups to be stored.
	Settings map[string]any `yaml:"settings"` // A map of key-value pairs for target-specific settings. TODO: Consider using BackupTargetSettings interface for type safety after implementing custom YAML unmarshaling.
}

// BackupScheduleConfig defines a single backup schedule
type BackupScheduleConfig struct {
	Enabled  bool   `yaml:"enabled"`  // If true, this specific schedule is active and backups will be attempted at the defined interval. (Valid: true or false)
	Hour     int    `yaml:"hour"`     // The hour of the day when the backup is scheduled to run. (Valid range: 0-23, where 0 is midnight and 23 is 11 PM)
	Minute   int    `yaml:"minute"`   // The minute of the hour when the backup is scheduled to run. (Valid range: 0-59)
	Weekday  string `yaml:"weekday"`  // For weekly schedules, the day of the week. Accepts: "Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday" (case-insensitive), or numeric: "0" (Sunday) through "6" (Saturday). Empty or ignored for daily schedules.
	IsWeekly bool   `yaml:"isweekly"` // If true, this schedule is weekly (runs on the specified Weekday at Hour:Minute). If false, it's a daily schedule (runs every day at Hour:Minute). (Valid: true or false)
}

// BackupConfig contains backup-related configuration
type BackupConfig struct {
	Enabled        bool                   `yaml:"enabled"`         // Global flag to enable or disable the entire backup system. If false, no backups (manual or scheduled) will occur.
	Debug          bool                   `yaml:"debug"`           // If true, enables detailed debug logging for backup operations.
	Encryption     bool                   `yaml:"encryption"`      // If true, enables encryption for backup archives. Requires EncryptionKey to be set.
	EncryptionKey  string                 `yaml:"encryption_key"`  // Base64-encoded encryption key used for AES-256-GCM encryption of backup archives. Must be kept secret and safe.
	SanitizeConfig bool                   `yaml:"sanitize_config"` // If true, sensitive information (like passwords, API keys) will be removed from the configuration file copy that is included in the backup archive.
	Retention      BackupRetention        `yaml:"retention"`       // Defines policies for how long and how many backups are kept.
	Targets        []BackupTarget         `yaml:"targets"`         // A list of configured backup targets (destinations) where backup archives will be stored.
	Schedules      []BackupScheduleConfig `yaml:"schedules"`       // A list of schedules (e.g., daily, weekly) that define when automatic backups should run.

	// OperationTimeouts defines timeouts for various backup operations
	OperationTimeouts struct {
		Backup  time.Duration `yaml:"backup"`  // Maximum duration allowed for the entire backup operation for a single source (including data extraction, archiving, compression, encryption). Default: 2h.
		Store   time.Duration `yaml:"store"`   // Maximum duration allowed for storing a single backup archive to one target. Default: 15m.
		Cleanup time.Duration `yaml:"cleanup"` // Maximum duration allowed for the backup cleanup process (deleting old backups based on retention policy). Default: 10m.
		Delete  time.Duration `yaml:"delete"`  // Maximum duration allowed for deleting a single backup archive from a target. Default: 2m.
	}
}

// Settings contains all configuration options for the BirdNET-Go application.
type Settings struct {
	Debug bool // true to enable debug mode

	// Runtime values, not stored in config file
	Version            string   `yaml:"-"` // Version from build
	BuildDate          string   `yaml:"-"` // Build date from build
	SystemID           string   `yaml:"-"` // Unique system identifier for telemetry
	ValidationWarnings []string `yaml:"-"` // Configuration validation warnings for telemetry

	Main struct {
		Name      string    // name of BirdNET-Go node, can be used to identify source of notes
		TimeAs24h bool      // true 24-hour time format, false 12-hour time format
		Log       LogConfig // logging configuration
	}

	BirdNET BirdNETConfig // BirdNET configuration

	Input InputConfig `yaml:"-"` // Input configuration for file and directory analysis

	Realtime  RealtimeSettings  // Realtime processing settings
	WebServer WebServerSettings // web server configuration
	Security  Security          // security configuration
	Sentry    SentrySettings    // Sentry error tracking configuration

	Output struct {
		File struct {
			Enabled bool   `yaml:"-"` // true to enable file output
			Path    string `yaml:"-"` // directory to output results
			Type    string `yaml:"-"` // table, csv
		}

		SQLite struct {
			Enabled bool   // true to enable sqlite output
			Path    string // path to sqlite database
		}

		MySQL struct {
			Enabled  bool   // true to enable mysql output
			Username string // username for mysql database
			Password string // password for mysql database
			Database string // database name for mysql database
			Host     string // host for mysql database
			Port     string // port for mysql database
		}
	}

	Backup BackupConfig // Backup configuration
}

// LogConfig defines the configuration for a log file
type LogConfig struct {
	Enabled     bool         // true to enable this log
	Path        string       // Path to the log file
	Rotation    RotationType // Type of log rotation
	MaxSize     int64        // Max size in bytes for RotationSize
	RotationDay string       // Day of the week for RotationWeekly (as a string: "Sunday", "Monday", etc.)
}

// RotationType defines different types of log rotations.
type RotationType string

const (
	RotationDaily  RotationType = "daily"
	RotationWeekly RotationType = "weekly"
	RotationSize   RotationType = "size"
)

// settingsInstance is the current settings instance
var (
	settingsInstance *Settings
	once             sync.Once
	settingsMutex    sync.RWMutex
)

// Load reads the configuration file and environment variables into GlobalConfig.
func Load() (*Settings, error) {
	settingsMutex.Lock()
	defer settingsMutex.Unlock()

	// Create a new settings struct
	settings := &Settings{}

	// Initialize viper and read config
	if err := initViper(); err != nil {
		return nil, errors.New(err).
			Category(errors.CategoryConfiguration).
			Context("operation", "init-viper").
			Build()
	}

	// Unmarshal the config into settings
	if err := viper.Unmarshal(settings); err != nil {
		return nil, errors.New(err).
			Category(errors.CategoryConfiguration).
			Context("operation", "unmarshal-config").
			Build()
	}

	// Validate settings
	if err := ValidateSettings(settings); err != nil {
		// Check if it's just a validation warning (contains fallback info)
		var validationErr ValidationError
		if errors.As(err, &validationErr) {
			// Report configuration issues to telemetry for debugging
			for _, errMsg := range validationErr.Errors {
				if strings.Contains(errMsg, "fallback") || strings.Contains(errMsg, "not supported") {
					// This is a warning about locale fallback - report to telemetry but don't fail
					log.Printf("Configuration warning: %s", errMsg)
					// Store the warning for later telemetry reporting
					settings.ValidationWarnings = append(settings.ValidationWarnings, errMsg)
					// Note: Telemetry reporting will happen later in birdnet package when Sentry is initialized
				} else {
					// This is a real validation error - fail the config load
					return nil, errors.New(err).
						Category(errors.CategoryValidation).
						Context("component", "settings").
						Context("error_msg", errMsg).
						Build()
				}
			}
		} else {
			// Other validation errors should fail the config load
			return nil, errors.New(err).
				Category(errors.CategoryValidation).
				Context("component", "settings").
				Build()
		}
	}

	// Log the loaded species settings for debugging
	/*
		log.Printf("Loaded Species Settings: Include: %v, Exclude: %v, Threshold: %v",
			settings.Realtime.Species.Include,
			settings.Realtime.Species.Exclude,
			settings.Realtime.Species.Threshold)
	*/

	// Save settings instance
	settingsInstance = settings
	return settingsInstance, nil
}

// initViper initializes viper with default values and reads the configuration file.
func initViper() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Get OS specific config paths
	configPaths, err := GetDefaultConfigPaths()
	if err != nil {
		return errors.New(err).
			Category(errors.CategoryConfiguration).
			Context("operation", "get-config-paths").
			Build()
	}

	// Assign config paths to Viper
	for _, path := range configPaths {
		viper.AddConfigPath(path)
	}

	// Set default values for each configuration parameter
	// function defined in defaults.go
	setDefaultConfig()

	// Read configuration file
	err = viper.ReadInConfig()
	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			// Config file not found, create config with defaults
			return createDefaultConfig()
		}
		// Report critical config file read errors
		return errors.New(err).
			Category(errors.CategoryFileIO).
			Context("operation", "read-config-file").
			Build()
	}

	return nil
}

// createDefaultConfig creates a default config file and writes it to the default config path
func createDefaultConfig() error {
	configPaths, err := GetDefaultConfigPaths()
	if err != nil {
		return errors.New(err).
			Category(errors.CategoryConfiguration).
			Context("operation", "create-default-config-paths").
			Build()
	}
	configPath := filepath.Join(configPaths[0], "config.yaml")
	defaultConfig := getDefaultConfig()

	// If the basicauth secret is not set, generate a random one
	if viper.GetString("security.basicauth.clientsecret") == "" {
		viper.Set("security.basicauth.clientsecret", GenerateRandomSecret())
	}

	// Create directories for config file
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return errors.New(err).
			Category(errors.CategoryFileIO).
			Context("operation", "create-config-dirs").
			Context("path", filepath.Dir(configPath)).
			Build()
	}

	// Write default config file with secure permissions (0600)
	// Only the owner should be able to read/write the config file for security
	if err := os.WriteFile(configPath, []byte(defaultConfig), 0o600); err != nil {
		return errors.New(err).
			Category(errors.CategoryFileIO).
			Context("operation", "write-default-config").
			Context("path", configPath).
			Build()
	}

	fmt.Println("Created default config file at:", configPath)
	return viper.ReadInConfig()
}

// getDefaultConfig reads the default configuration from the embedded config.yaml file.
func getDefaultConfig() string {
	data, err := fs.ReadFile(configFiles, "config.yaml")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
	return string(data)
}

// GetSettings returns the current settings instance
func GetSettings() *Settings {
	settingsMutex.RLock()
	defer settingsMutex.RUnlock()
	return settingsInstance
}

// SaveSettings saves the current settings to the configuration file.
// It uses UpdateYAMLConfig to handle the atomic write process.
func SaveSettings() error {
	settingsMutex.RLock()
	defer settingsMutex.RUnlock()

	// Create a deep copy of the settings
	settingsCopy := *settingsInstance

	// Create a separate copy of the species list
	speciesListMutex.RLock()
	settingsCopy.BirdNET.RangeFilter.Species = make([]string, len(settingsInstance.BirdNET.RangeFilter.Species))
	copy(settingsCopy.BirdNET.RangeFilter.Species, settingsInstance.BirdNET.RangeFilter.Species)
	speciesListMutex.RUnlock()

	// Find the path of the current config file
	configPath, err := FindConfigFile()
	if err != nil {
		return errors.New(err).
			Category(errors.CategoryFileIO).
			Context("operation", "find-config-file").
			Build()
	}

	// Save the settings to the config file
	if err := SaveYAMLConfig(configPath, &settingsCopy); err != nil {
		return errors.New(err).
			Category(errors.CategoryFileIO).
			Context("operation", "save-yaml-config").
			Context("path", configPath).
			Build()
	}

	log.Printf("Settings saved successfully to %s", configPath)
	return nil
}

// Setting returns the current settings instance, initializing it if necessary
func Setting() *Settings {
	once.Do(func() {
		if settingsInstance == nil {
			_, err := Load()
			if err != nil {
				// Fatal error loading settings - application cannot continue
				enhancedErr := errors.New(err).
					Category(errors.CategoryConfiguration).
					Context("operation", "load-settings-init").
					Build()
				log.Fatalf("Error loading settings: %v", enhancedErr)
			}
		}
	})
	return GetSettings()
}

// SaveYAMLConfig updates the YAML configuration file with new settings.
// It overwrites the existing file, not preserving comments or structure.
func SaveYAMLConfig(configPath string, settings *Settings) error {
	// Marshal the settings struct to YAML
	yamlData, err := yaml.Marshal(settings)
	if err != nil {
		return errors.New(err).
			Category(errors.CategoryConfiguration).
			Context("operation", "yaml-marshal").
			Build()
	}

	// Write the YAML data to a temporary file
	// This is done to ensure atomic write operation
	tempFile, err := os.CreateTemp(filepath.Dir(configPath), "config-*.yaml")
	if err != nil {
		return errors.New(err).
			Category(errors.CategoryFileIO).
			Context("operation", "create-temp-file").
			Context("dir", filepath.Dir(configPath)).
			Build()
	}
	tempFileName := tempFile.Name()
	// Ensure the temporary file is removed in case of any failure
	defer func() {
		if err := os.Remove(tempFileName); err != nil && !os.IsNotExist(err) {
			log.Printf("Failed to remove temporary file: %v", err)
		}
	}()

	// Write the YAML data to the temporary file
	if _, err := tempFile.Write(yamlData); err != nil {
		// Best effort close on error path
		_ = tempFile.Close()
		return errors.New(err).
			Category(errors.CategoryFileIO).
			Context("operation", "write-temp-file").
			Build()
	}
	// Close the temporary file after writing
	if err := tempFile.Close(); err != nil {
		return errors.New(err).
			Category(errors.CategoryFileIO).
			Context("operation", "close-temp-file").
			Build()
	}

	// Try to rename the temporary file to replace the original config file
	// This is typically an atomic operation on most filesystems
	if err := os.Rename(tempFileName, configPath); err != nil {
		// If rename fails (e.g., cross-device link), fall back to copy & delete
		// This might happen when the temp directory is on a different filesystem
		if err := moveFile(tempFileName, configPath); err != nil {
			return errors.New(err).
				Category(errors.CategoryFileIO).
				Context("operation", "move-config-file").
				Context("src", tempFileName).
				Context("dst", configPath).
				Build()
		}
	}

	// If we've reached this point, the operation was successful
	return nil
}

// GenerateRandomSecret generates a URL-safe base64 encoded random string
// suitable for use as a client secret. The output is 43 characters long,
// providing 256 bits of entropy.
func GenerateRandomSecret() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Log the error and return a safe fallback or empty string
		enhancedErr := errors.New(err).
			Category(errors.CategorySystem).
			Context("operation", "generate-random-secret").
			Build()
		log.Printf("Failed to generate random secret: %v", enhancedErr)
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

// GetWeatherSettings returns the appropriate weather settings based on the configuration
func (s *Settings) GetWeatherSettings() (provider string, openweather OpenWeatherSettings) {
	// First check new format
	if s.Realtime.Weather.Provider != "" {
		return s.Realtime.Weather.Provider, s.Realtime.Weather.OpenWeather
	}

	if s.Realtime.OpenWeather.Enabled {
		return "openweather", s.Realtime.OpenWeather
	}

	// Default to YrNo if nothing is configured
	return "yrno", OpenWeatherSettings{}
}
