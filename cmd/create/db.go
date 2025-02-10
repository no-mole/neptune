package create

import (
	"bytes"
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

func initModelFile(baseDir string, modelTemplate map[string][]byte, modName string) error {
	config, err := readDBConfig()
	if err != nil && err.Error() == "not found db config" {
		return nil
	} else if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User, config.Password, config.Host, config.Port, config.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}

	// 获取所有表名
	var tables []string
	err = db.Raw("SHOW TABLES").Scan(&tables).Error
	if err != nil {
		return fmt.Errorf("获取表名失败: %v", err)
	}

	// 为每个表生成模型文件
	for _, table := range tables {
		if err := generateModelFiles(db, baseDir, table, modName, modelTemplate); err != nil {
			return fmt.Errorf("生成表 %s 的模型文件失败: %v", table, err)
		}
	}

	// 创建外层的model
	file, err := os.Create(path.Join(baseDir, "model", "model.go")) // 创建文件
	if err != nil {
		return fmt.Errorf("生成外层model失败: %v", err)
	}
	defer file.Close() // 关闭文件

	fileInfo := `package model

const (
	MysqlEngine = "mysql"
	RedisEngine = "redis"
	MongoEngine = "mongo"
)
`
	_, err = file.WriteString(fileInfo) // 写入内容
	if err != nil {
		return fmt.Errorf("写入失败: %v", err)
	}
	return nil
}

func readDBConfig() (*DBConfig, error) {
	var configFile string

	// 尝试读取 db.yaml 或 db.yml
	if _, err := os.Stat("db.yaml"); err == nil {
		configFile = "db.yaml"
	} else if _, err := os.Stat("db.yml"); err == nil {
		configFile = "db.yml"
	} else {
		return nil, fmt.Errorf("not found db config")
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config DBConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

type FieldInfo struct {
	Name     string
	Type     string
	Column   string
	JSONName string
	ModName  string
}

type TemplateData struct {
	TableName  string
	StructName string
	ModName    string
	Fields     []FieldInfo
}

func generateModelFiles(db *gorm.DB, baseDir, tableName, modName string, modelTemplate map[string][]byte) error {
	// 创建目录
	modelDir := filepath.Join(baseDir, "model", tableName)
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return err
	}

	// 获取表结构信息
	var columns []struct {
		Field   string `gorm:"column:Field"`
		Type    string `gorm:"column:Type"`
		Null    string `gorm:"column:Null"`
		Key     string `gorm:"column:Key"`
		Default string `gorm:"column:Default"`
		Extra   string `gorm:"column:Extra"`
	}

	db.Raw(fmt.Sprintf("SHOW COLUMNS FROM %s", tableName)).Scan(&columns)

	// 准备模板数据
	data := TemplateData{
		TableName:  tableName,
		StructName: toCamelCase(tableName),
		ModName:    modName,
		Fields:     make([]FieldInfo, 0, len(columns)),
	}

	for _, col := range columns {
		jsonName := strings.ToLower(col.Field)
		data.Fields = append(data.Fields, FieldInfo{
			Name:     toCamelCase(col.Field),
			Type:     mysqlTypeToGo(col.Type),
			Column:   col.Field,
			JSONName: jsonName,
		})
	}

	for fileName, fileBody := range modelTemplate {
		name := strings.Trim(strings.TrimSuffix(strings.TrimPrefix(fileName, "template"), "template"), "/")
		outputPath := path.Join(baseDir, "model", tableName, name)

		err := generateFileFromTemplate(name, fileBody, outputPath, data)
		if err != nil {
			return err
		}
	}

	return nil
}

func generateFileFromTemplate(fileName string, fileBody []byte, outputPath string, data interface{}) error {
	// 解析模板
	tmpl, err := template.New(fileName).Parse(string(fileBody))
	if err != nil {
		return fmt.Errorf("解析模板失败: %v", err)
	}

	// 执行模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("执行模板失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

func toCamelCase(s string) string {
	// 创建一个 title caser
	caser := cases.Title(language.English)
	parts := strings.Split(s, "_")
	for i := range parts {
		parts[i] = caser.String(parts[i])
	}
	return strings.Join(parts, "")
}

func mysqlTypeToGo(mysqlType string) string {
	mysqlType = strings.ToLower(mysqlType)
	switch {
	case strings.Contains(mysqlType, "int"):
		return "int64"
	case strings.Contains(mysqlType, "char"), strings.Contains(mysqlType, "text"):
		return "string"
	case strings.Contains(mysqlType, "datetime"), strings.Contains(mysqlType, "timestamp"):
		return "string"
	case strings.Contains(mysqlType, "decimal"), strings.Contains(mysqlType, "float"):
		return "float64"
	default:
		return "string"
	}
}
