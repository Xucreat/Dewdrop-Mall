package util

/*添加日志分级存储和日志轮转*/

import (
	"log"
	"os"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

var LogrusObj *logrus.Logger

// 初始化 Logrus 日志
func InitLog() {
	if LogrusObj != nil {
		return
	}

	// 实例化 Logrus
	logger := logrus.New()

	// 设置日志格式
	logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// 设置一般日志的轮转
	logger.SetOutput(&lumberjack.Logger{
		Filename:   getLogFilePath() + "app.log",
		MaxSize:    10,   // 最大日志文件大小 (MB)
		MaxBackups: 5,    // 保留旧文件的最大个数
		MaxAge:     30,   // 保留旧文件的最长天数
		Compress:   true, // 是否压缩/归档旧文件
	})

	// 设置日志级别
	logger.SetLevel(logrus.InfoLevel)

	// 创建单独的错误日志轮转
	errorLogFile := &lumberjack.Logger{
		Filename:   getLogFilePath() + "error.log",
		MaxSize:    10,   // 最大日志文件大小 (MB)
		MaxBackups: 5,    // 保留旧文件的最大个数
		MaxAge:     30,   // 保留旧文件的最长天数
		Compress:   true, // 是否压缩/归档旧文件
	}

	// 添加错误级别日志 Hook
	logger.AddHook(&LogLevelHook{
		Writer:    errorLogFile,
		LogLevels: []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel},
	})

	LogrusObj = logger
}

// 获取日志文件路径
func getLogFilePath() string {
	logFilePath := ""
	if dir, err := os.Getwd(); err == nil {
		logFilePath = dir + "/logs/"
	}
	_, err := os.Stat(logFilePath)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(logFilePath, 0777); err != nil {
			log.Println("Error creating log directory:", err)
		}
	}
	return logFilePath
}

// LogLevelHook 定义日志级别的 Hook
type LogLevelHook struct {
	Writer    *lumberjack.Logger
	LogLevels []logrus.Level
}

// Fire 将日志写入到指定的日志文件
func (hook *LogLevelHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write([]byte(line))
	return err
}

// Levels 返回应该写入的日志级别
func (hook *LogLevelHook) Levels() []logrus.Level {
	return hook.LogLevels
}

// GormLogWriter 实现 gorm 的 logger.Writer 接口
type GormLogWriter struct {
	logrus *logrus.Logger
}

// NewGormLogWriter 创建新的 GormLogWriter
func NewGormLogWriter(logger *logrus.Logger) *GormLogWriter {
	if logger == nil {
		// 防止 logrus 未初始化
		panic("LogrusObj is nil. Please ensure InitLog is called before creating GormLogWriter.")
	}

	return &GormLogWriter{
		logrus: logger,
	}
}

// Printf 实现 gorm logger.Writer 接口
// 通过 Printf 方法将 gorm 的日志信息转发到 Logrus
// Logrus 可以统一管理日志，无论是应用层日志还是数据库操作日志，都能通过同一个日志系统输出。
func (w *GormLogWriter) Printf(format string, args ...interface{}) {
	w.logrus.Infof(format, args...)
}
