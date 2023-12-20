package configFile

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

func (s *Service) newConfig(name string) *Config {
	c := &Config{
		service: s,
		name:    name,
	}
	c.groups = make(map[string]Object)
	return c
}

type Config struct {
	service *Service
	name    string
	groups  map[string]Object
}

func (c *Config) RootDirectory() string {
	return c.service.RootDirectory
}

func (c *Config) Save() error {
	return c.service.SaveConfigWithFileName(c.name)
}

func (c *Config) Load() error {
	loaded, err := c.service.LoadConfigWithFileName(c.name)
	if err != nil {
		return err
	}

	c.groups = loaded.groups

	return nil
}

func (c *Config) GroupKeys() []string {
	groups := make([]string, 0)

	for groupKey := range c.groups {
		groups = append(groups, groupKey)
	}

	return groups
}

func (c *Config) Group(s string) Object {
	if c.groups == nil {
		c.groups = make(map[string]Object)
	}

	if _, found := c.groups[s]; !found {
		c.groups[s] = Object{}
	}

	return c.groups[s]
}

func (c *Config) Marshal() ([]byte, error) {
	return json.MarshalIndent(c.groups, "", "  ")
}

func (c *Config) Unmarshal(data []byte) error {
	return json.Unmarshal(data, &c.groups)
}

type Object map[string]interface{}

func (n Object) Marshal() ([]byte, error) {
	return json.Marshal(n)
}

func (n Object) Unmarshal(data []byte) error {
	return json.Unmarshal(data, n)
}

func (n Object) Keys() (keys []string) {
	for key := range n {
		keys = append(keys, key)
	}

	return keys
}

func (n Object) Set(key string, value interface{}) {
	n[key] = value
}

func (n Object) Get(key string) interface{} {
	if n == nil {
		return nil
	}

	return n[key]
}

func (n Object) GetString(key string) string {
	if n == nil {
		return ""
	}

	value, _ := n[key].(string)
	return value
}

func (n Object) GetStrings(key string) []string {
	if n == nil {
		return nil
	}

	res := make([]string, 0)

	data, _ := json.Marshal(n[key])
	json.Unmarshal(data, &res)

	return res
}

func (n Object) GetInt(key string) int {
	if n == nil {
		return 0
	}

	if value, ok := n[key].(int); ok {
		return value
	}

	val, err := strconv.ParseInt(fmt.Sprintf("%v", n[key]), 10, 32)
	if err != nil {
		return 0
	}

	return int(val)
}

func (n Object) GetBool(key string) bool {
	if n == nil {
		return false
	}

	value, _ := n[key].(bool)
	return value
}

func (n Object) GetDuration(key string) time.Duration {
	if n == nil {
		return 0
	}

	if d, err := time.ParseDuration(n.GetString(key)); err == nil {
		return d
	}

	return time.Second
}

func (n Object) GetJson(key string) []byte {
	if n == nil {
		return nil
	}

	data, _ := json.Marshal(n[key])

	return data
}

func (n Object) IsSet(key string) bool {
	_, ok := n[key]
	return ok
}

func (n Object) SetDefault(key string, defaultValue interface{}) {
	if _, ok := n[key]; !ok {
		n[key] = defaultValue
	}
}
