package biz

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-kratos/kratos/v2/log"
)

// PronunciationService 发音服务接口
type PronunciationService interface {
	GetAudioURLs(word string) []string
	LoadDictionary(filePath string) error
}

// pronunciationService 发音服务实现
type pronunciationService struct {
	dict map[string][]string
	mu   sync.RWMutex
	log  *log.Helper
}

// NewPronunciationService 创建发音服务
func NewPronunciationService(logger log.Logger) PronunciationService {
	service := &pronunciationService{
		dict: make(map[string][]string),
		log:  log.NewHelper(logger),
	}
	/*	if err := service.LoadDictionary("/build/ultimate.json"); err != nil {
		panic(err)

	}*/
	return service
}

// LoadDictionary 加载发音字典文件
func (p *pronunciationService) LoadDictionary(filePath string) error {
	// 如果是相对路径，尝试多个可能的路径（用于本地开发）
	if !filepath.IsAbs(filePath) {
		// 尝试多个可能的路径
		possiblePaths := []string{
			filepath.Join("backend", filePath),
			filepath.Join(".", filePath),
			filepath.Join("..", filePath),
			filePath, // 直接尝试（在 Docker 中可能是相对于工作目录）
		}

		var found bool
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				filePath = path
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("dictionary file not found, tried: %v", possiblePaths)
		}
	}

	p.log.Infof("Loading pronunciation dictionary from: %s", filePath)

	// 读取文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read dictionary file: %w", err)
	}

	// 解析 JSON
	var dict map[string][]string
	if err := json.Unmarshal(data, &dict); err != nil {
		return fmt.Errorf("failed to parse dictionary JSON: %w", err)
	}

	// 更新字典（加锁保护）
	p.mu.Lock()
	p.dict = dict
	p.mu.Unlock()

	p.log.Infof("Successfully loaded %d words from pronunciation dictionary", len(dict))
	return nil
}

// GetAudioURLs 获取单词的音频 URL 列表
func (p *pronunciationService) GetAudioURLs(word string) []string {
	if word == "" {
		return nil
	}

	// 转换为小写进行查找（不区分大小写）
	wordLower := strings.ToLower(strings.TrimSpace(word))

	p.mu.RLock()
	defer p.mu.RUnlock()

	// 直接查找
	if urls, ok := p.dict[wordLower]; ok && len(urls) > 0 {
		return urls
	}

	// 如果直接查找失败，尝试查找变体
	// 1. 查找带引号的变体（如 'hood）
	if urls, ok := p.dict["'"+wordLower]; ok && len(urls) > 0 {
		return urls
	}

	// 2. 查找去掉标点符号的版本
	wordClean := strings.Trim(wordLower, ".,!?;:'\"()[]{}")
	if wordClean != wordLower {
		if urls, ok := p.dict[wordClean]; ok && len(urls) > 0 {
			return urls
		}
	}

	// 3. 尝试查找首字母大写的版本
	wordTitle := strings.Title(wordLower)
	if wordTitle != wordLower {
		if urls, ok := p.dict[wordTitle]; ok && len(urls) > 0 {
			return urls
		}
	}

	return nil
}
