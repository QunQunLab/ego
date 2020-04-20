package conf

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	CRLF     = '\n'
	Delimit  = ","
	Split    = "="
	Comment  = "#"
	SectionB = "["
	SectionE = "]"
	Include  = "include"
)

//	Section
//
//	# common config have no sector
//	commonKey commonVal
//	commonKey commonVal1,commonVal2
//
///	# include other file
//	include ../common
//
//	[sector]
//	sectorKey sectorVal1,sectorVal2
type Section struct {
	delimit string
	sector  string
	val     map[string]string // key val1,val2
}

// An NoKeyError describes a key that was not found in the section.
type NoKeyError struct {
	Key     string
	Section string
}

func (e *NoKeyError) Error() string {
	return fmt.Sprintf("key: \"%s\" not found in [%s]", e.Key, e.Section)
}

// String get config string value.
func (s *Section) String(key string) (string, error) {
	if v, ok := s.val[key]; ok {
		return v, nil
	} else {
		return "", &NoKeyError{Key: key, Section: s.sector}
	}
}

// Strings get config []string value.
func (s *Section) Strings(key string) ([]string, error) {
	if v, ok := s.val[key]; ok {
		return strings.Split(v, s.delimit), nil
	} else {
		return nil, &NoKeyError{Key: key, Section: s.sector}
	}
}

// Int get config int value.
func (s *Section) Int(key string) (int64, error) {
	if v, ok := s.val[key]; ok {
		return strconv.ParseInt(v, 10, 64)
	} else {
		return 0, &NoKeyError{Key: key, Section: s.sector}
	}
}

// Uint get config uint value.
func (s *Section) Uint(key string) (uint64, error) {
	if v, ok := s.val[key]; ok {
		return strconv.ParseUint(v, 10, 64)
	} else {
		return 0, &NoKeyError{Key: key, Section: s.sector}
	}
}

// Float get config float value.
func (s *Section) Float(key string) (float64, error) {
	if v, ok := s.val[key]; ok {
		return strconv.ParseFloat(v, 64)
	} else {
		return 0, &NoKeyError{Key: key, Section: s.sector}
	}
}

// Bool get config boolean value.
// "yes", "1", "y", "true", "enable" means true.
// "no", "0", "n", "false", "disable" means false.
// if the specified value unknown then return false.
func (s *Section) Bool(key string) (bool, error) {
	if v, ok := s.val[key]; ok {
		return parseBool(strings.ToLower(v)), nil
	} else {
		return false, &NoKeyError{Key: key, Section: s.sector}
	}
}

// MemSize Byte get config byte number value.
// 1kb = 1k = 1024.
// 1mb = 1m = 1024 * 1024.
// 1gb = 1g = 1024 * 1024 * 1024.
func (s *Section) MemSize(key string) (int, error) {
	if v, ok := s.val[key]; ok {
		return parseMemory(v)
	} else {
		return 0, &NoKeyError{Key: key, Section: s.sector}
	}
}

// Duration get config second value.
// 1s = 1sec = 1.
// 1m = 1min = 60.
// 1h = 1hour = 60 * 60.
func (s *Section) Duration(key string) (time.Duration, error) {
	if v, ok := s.val[key]; ok {
		if t, err := parseTime(v); err != nil {
			return 0, err
		} else {
			return time.Duration(t), nil
		}
	} else {
		return 0, &NoKeyError{Key: key, Section: s.sector}
	}
}

func parseTime(v string) (int64, error) {
	unit := int64(time.Nanosecond)
	subIdx := len(v)
	if strings.HasSuffix(v, "ms") {
		unit = int64(time.Millisecond)
		subIdx = subIdx - 2
	} else if strings.HasSuffix(v, "s") {
		unit = int64(time.Second)
		subIdx = subIdx - 1
	} else if strings.HasSuffix(v, "sec") {
		unit = int64(time.Second)
		subIdx = subIdx - 3
	} else if strings.HasSuffix(v, "m") {
		unit = int64(time.Minute)
		subIdx = subIdx - 1
	} else if strings.HasSuffix(v, "min") {
		unit = int64(time.Minute)
		subIdx = subIdx - 3
	} else if strings.HasSuffix(v, "h") {
		unit = int64(time.Hour)
		subIdx = subIdx - 1
	} else if strings.HasSuffix(v, "hour") {
		unit = int64(time.Hour)
		subIdx = subIdx - 4
	}
	b, err := strconv.ParseInt(v[:subIdx], 10, 64)
	if err != nil {
		return 0, err
	}
	return b * unit, nil
}

// Keys return all the section keys.
func (s *Section) Keys() []string {
	var keys []string
	for k := range s.val {
		keys = append(keys, k)
	}
	return keys
}

// Config config
type Config struct {
	// common config
	Common map[string]string // commonKey commonVal1,commonVal2

	// sectors
	Sector map[string]*Section

	// config file path
	File string

	// default config
	Comment string
	Split   string
	Delimit string
}

// New return a new default Config object (Comment = '#', Split = ' ', Delimit = ',')
func New() *Config {
	return &Config{
		Common: make(map[string]string),
		Sector: make(map[string]*Section),

		// default config
		Comment: Comment,
		Split:   Split,
		Delimit: Delimit,
	}
}

// ParseReader parse from io.Reader
func (c *Config) ParseReader(reader io.Reader) error {
	var (
		line      int
		r         = bufio.NewReader(reader)
		sector    *Section
		sectorKey string
		key       string
		val       string
	)
	for {
		// process include
		// process common config
		// process sector
		line++
		row, err := r.ReadString(CRLF)
		if err != nil && err != io.EOF {
			return err
		}

		if err == io.EOF && len(row) == 0 {
			break
		}
		row = strings.TrimSpace(row)

		// process comment
		if len(row) == 0 || strings.HasPrefix(row, c.Comment) {
			// comment or empty line
			continue
		}
		if strings.HasPrefix(row, SectionB) {
			if !strings.HasSuffix(row, SectionE) {
				return fmt.Errorf("no end sector %s at file:%v line:%v", SectionE, c.File, line)
			}

			sectorKey = row[1 : len(row)-1]
			if _, ok := c.Sector[sectorKey]; ok {
				return fmt.Errorf("sector key %v already exists at file:%v line:%v", sectorKey, c.File, line)
			} else {
				sector = &Section{
					delimit: c.Delimit,
					sector:  sectorKey,
					val:     make(map[string]string),
				}
				c.Sector[sectorKey] = sector
			}
			continue
		}

		// key/val in a row
		idx := strings.Index(row, c.Split)
		if idx > 0 {
			key = strings.TrimSpace(row[:idx])
			if len(row) > idx {
				val = strings.TrimSpace(row[idx+1:])
			}
		} else {
			//return fmt.Errorf("no split in key row %v at file:%v line:%v", row, c.File, line)
		}

		if sector == nil {
			// process include
			if strings.Contains(row, Include) {
				includes := strings.SplitN(row, " ", 2)
				fileName := strings.Trim(includes[1], c.Split)
				abs, _ := filepath.Abs(c.File)
				file := path.Join(path.Dir(abs), fileName)
				if err = c.Parse(file); err != nil {
					return err
				}
			}
			// process common config
			if _, ok := c.Common[key]; ok {
				return fmt.Errorf("same common key %v at file:%v line:%v", key, c.File, line)
			}
			c.Common[key] = val
		} else {
			if c.Sector[sectorKey].val == nil {
				c.Sector[sectorKey].val = make(map[string]string)
			}
			if _, ok := c.Sector[sectorKey].val[key]; ok {
				return fmt.Errorf("section %s already has key: %s at file:%v line:%d", sectorKey, key, c.File, line)
			} else {
				c.Sector[sectorKey].val[key] = val
			}
		}
	}

	return nil
}

// Parse parse file
func (c *Config) Parse(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	c.File = file
	return c.ParseReader(f)
}

// Reload reload config
func (c *Config) Reload() (*Config, error) {
	nc := &Config{
		Common: make(map[string]string),
		Sector: make(map[string]*Section),
		File:   c.File,

		// config
		Comment: c.Comment,
		Split:   c.Split,
		Delimit: c.Delimit,
	}
	err := nc.Parse(c.File)
	if err != nil {
		return nil, err
	}
	return nc, nil
}

// Get get a config section by key.
func (c *Config) Get(section string) *Section {
	s, _ := c.Sector[section]
	return s
}

// GetKey get common key
func (c *Config) GetKey(key string) string {
	return c.Common[key]
}

// GetKeys get common key slice
func (c *Config) GetKeys(key string) []string {
	return strings.Split(c.Common[key], c.Delimit)
}

// Unmarshal unmarshal struct
// memory
const (
	Byte = 1
	KB   = 1024 * Byte
	MB   = 1024 * KB
	GB   = 1024 * MB
)

// timer
const (
	Second = 1
	Minute = 60 * Second
	Hour   = 60 * Minute
)

// An InvalidUnmarshalError describes an invalid argument passed to Unmarshal.
// (The argument to Unmarshal must be a non-nil pointer.)
type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "Unmarshal(nil)"
	}
	if e.Type.Kind() != reflect.Ptr {
		return "Unmarshal(non-pointer " + e.Type.String() + ")"
	}
	return "Unmarshal(nil " + e.Type.String() + ")"
}

// Unmarshal parses the flag struct and stores the result in the value
// pointed to by v.
//
// Struct values encode as flag objects. Each exported struct field
// becomes a member of the object unless
//   - the field's tag is "-", or
//   - the field is empty and its tag specifies the "omitempty" option.
// The empty values are false, 0, any
// nil pointer or interface value, and any array, slice, map, or string of
// length zero. The object's section and key string is the struct field name
// but can be specified in the struct field's tag value. The "flag" key in
// the struct field's tag value is the key name, followed by an optional comma
// and options. Examples:
//
//   // Field is ignored by this package.
//   Field int `flag:"-"`
//
//   // Field appears in flag section "base" as key "myName".
//   Field int `flag:"base:myName"`
//
//   // Field appears in flag section "base" as key "myName", the value split
//   // by delimiter ",".
//   Field []string `flag:"base:myName:,"`
//
//   // Field appears in flag section "base" as key "myName", the value split
//   // by delimiter "," and key-value is splited by "=".
//   Field map[int]string `flag:"base:myName:,"`
//
//   // Field appears in flag section "base" as key "myName", the value
//   // convert to time.Duration. When has extra tag "time", then flag can
//   // parse such "1h", "1s" config values.
//   //
//   // Note the extra tag "time" only effect the int64 (time.Duration is int64)
//   Field time.Duration `flag:"base:myName:time"`
//
//   // Field appears in flag section "base" as key "myName", when has extra
//   // tag, then flag can parse like "1gb", "1mb" config values.
//   //
//   // Note the extra tag "memory" only effect the int (memory size is int).
//   Field int `flag:"base:myName:memory"`
//
func (c *Config) Unmarshal(v interface{}, flag string) error {
	vv := reflect.ValueOf(v)
	if vv.Kind() != reflect.Ptr || vv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}
	rv := vv.Elem()
	rt := rv.Type()
	n := rv.NumField()
	// enum every struct field
	for i := 0; i < n; i++ {
		vf := rv.Field(i)
		tf := rt.Field(i)
		tag := tf.Tag.Get(flag)
		// if tag empty or "-" ignore
		if tag == "-" || tag == "" || tag == "omitempty" {
			continue
		}
		tagArr := strings.SplitN(tag, ":", 3)
		if len(tagArr) < 2 {
			return errors.New(fmt.Sprintf("error tag: %s, must be section:field:delim(optional)", tag))
		}
		section := tagArr[0]
		key := tagArr[1]
		s := c.Get(section)
		if s == nil {
			// no config section
			continue
		}
		value, ok := s.val[key]
		if !ok {
			// no config key
			continue
		}
		err := c.getVal(&vf, &tf, tagArr, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) UnmarshalSection(v interface{}, section, flag string) error {
	vv := reflect.ValueOf(v)
	if vv.Kind() != reflect.Ptr || vv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}
	rv := vv.Elem()
	rt := rv.Type()
	n := rv.NumField()
	// enum every struct field
	for i := 0; i < n; i++ {
		vf := rv.Field(i)
		tf := rt.Field(i)
		tag := tf.Tag.Get(flag)
		// if tag empty or "-" ignore
		if tag == "-" || tag == "" || tag == "omitempty" {
			continue
		}
		tagArr := strings.SplitN(tag, ":", 3)
		if len(tagArr) < 2 {
			return errors.New(fmt.Sprintf("error tag: %s, must be section:field:delim(optional)", tag))
		}
		section := tagArr[0]
		key := tagArr[1]
		if section != section {
			continue
		}
		s := c.Get(section)
		if s == nil {
			// no config section
			continue
		}
		value, ok := s.val[key]
		if !ok {
			// no config key
			continue
		}
		err := c.getVal(&vf, &tf, tagArr, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) getVal(vf *reflect.Value, tf *reflect.StructField, tagArr []string, value string) error {
	switch vf.Kind() {
	case reflect.String:
		vf.SetString(value)
	case reflect.Bool:
		vf.SetBool(parseBool(value))
	case reflect.Float32:
		if tmp, err := strconv.ParseFloat(value, 32); err != nil {
			return err
		} else {
			vf.SetFloat(tmp)
		}
	case reflect.Float64:
		if tmp, err := strconv.ParseFloat(value, 64); err != nil {
			return err
		} else {
			vf.SetFloat(tmp)
		}
	case reflect.Int:
		if len(tagArr) == 3 {
			format := tagArr[2]
			// parse memory size
			if format == "memory" {
				if tmp, err := parseMemory(value); err != nil {
					return err
				} else {
					vf.SetInt(int64(tmp))
				}
			} else {
				return errors.New(fmt.Sprintf("unknown tag: %s in struct field: %s (support tags: \"memory\")", format, tf.Name))
			}
		} else {
			if tmp, err := strconv.ParseInt(value, 10, 32); err != nil {
				return err
			} else {
				vf.SetInt(tmp)
			}
		}
	case reflect.Int8:
		if tmp, err := strconv.ParseInt(value, 10, 8); err != nil {
			return err
		} else {
			vf.SetInt(tmp)
		}
	case reflect.Int16:
		if tmp, err := strconv.ParseInt(value, 10, 16); err != nil {
			return err
		} else {
			vf.SetInt(tmp)
		}
	case reflect.Int32:
		if tmp, err := strconv.ParseInt(value, 10, 32); err != nil {
			return err
		} else {
			vf.SetInt(tmp)
		}
	case reflect.Int64:
		if len(tagArr) == 3 {
			format := tagArr[2]
			// parse time
			if format == "time" {
				if tmp, err := parseTime(value); err != nil {
					return err
				} else {
					vf.SetInt(tmp)
				}
			} else {
				return errors.New(fmt.Sprintf("unknown tag: %s in struct field: %s (support tags: \"time\")", format, tf.Name))
			}
		} else {
			if tmp, err := strconv.ParseInt(value, 10, 64); err != nil {
				return err
			} else {
				vf.SetInt(tmp)
			}
		}
	case reflect.Uint:
		if tmp, err := strconv.ParseUint(value, 10, 32); err != nil {
			return err
		} else {
			vf.SetUint(tmp)
		}
	case reflect.Uint8:
		if tmp, err := strconv.ParseUint(value, 10, 8); err != nil {
			return err
		} else {
			vf.SetUint(tmp)
		}
	case reflect.Uint16:
		if tmp, err := strconv.ParseUint(value, 10, 16); err != nil {
			return err
		} else {
			vf.SetUint(tmp)
		}
	case reflect.Uint32:
		if tmp, err := strconv.ParseUint(value, 10, 32); err != nil {
			return err
		} else {
			vf.SetUint(tmp)
		}
	case reflect.Uint64:
		if tmp, err := strconv.ParseUint(value, 10, 64); err != nil {
			return err
		} else {
			vf.SetUint(tmp)
		}
	case reflect.Slice:
		delim := ","
		if len(tagArr) > 2 {
			delim = tagArr[2]
		}
		strs := strings.Split(value, delim)
		sli := reflect.MakeSlice(tf.Type, 0, len(strs))
		for _, str := range strs {
			vv, err := getValue(tf.Type.Elem().String(), str)
			if err != nil {
				return err
			}
			sli = reflect.Append(sli, vv)
		}
		vf.Set(sli)
	case reflect.Map:
		delim := ","
		if len(tagArr) > 2 {
			delim = tagArr[2]
		}
		strs := strings.Split(value, delim)
		m := reflect.MakeMap(tf.Type)
		for _, str := range strs {
			mapStrs := strings.SplitN(str, "=", 2)
			if len(mapStrs) < 2 {
				return errors.New(fmt.Sprintf("error map: %s, must be split by \"=\"", str))
			}
			vk, err := getValue(tf.Type.Key().String(), mapStrs[0])
			if err != nil {
				return err
			}
			vv, err := getValue(tf.Type.Elem().String(), mapStrs[1])
			if err != nil {
				return err
			}
			m.SetMapIndex(vk, vv)
		}
		vf.Set(m)
	default:
		return errors.New(fmt.Sprintf("cannot unmarshall unsuported kind: %s into struct field: %s", vf.Kind().String(), tf.Name))
	}
	return nil
}

// getValue parse String to the type "t" reflect.Value.
func getValue(t, v string) (reflect.Value, error) {
	var vv reflect.Value
	switch t {
	case "bool":
		d := parseBool(v)
		vv = reflect.ValueOf(d)
	case "int":
		d, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(int(d))
	case "int8":
		d, err := strconv.ParseInt(v, 10, 8)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(int8(d))
	case "int16":
		d, err := strconv.ParseInt(v, 10, 16)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(int16(d))
	case "int32":
		d, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(int32(d))
	case "int64":
		d, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(int64(d))
	case "uint":
		d, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(uint(d))
	case "uint8":
		d, err := strconv.ParseUint(v, 10, 8)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(uint8(d))
	case "uint16":
		d, err := strconv.ParseUint(v, 10, 16)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(uint16(d))
	case "uint32":
		d, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(uint32(d))
	case "uint64":
		d, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(uint64(d))
	case "float32":
		d, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(float32(d))
	case "float64":
		d, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(float64(d))
	case "string":
		vv = reflect.ValueOf(v)
	default:
		return vv, errors.New(fmt.Sprintf("unkown type: %s", t))
	}
	return vv, nil
}

func parseBool(v string) bool {
	if v == "true" || v == "yes" || v == "1" || v == "y" || v == "enable" {
		return true
	} else if v == "false" || v == "no" || v == "0" || v == "n" || v == "disable" {
		return false
	} else {
		return false
	}
}

func parseMemory(v string) (int, error) {
	unit := Byte
	subIdx := len(v)
	if strings.HasSuffix(v, "k") {
		unit = KB
		subIdx = subIdx - 1
	} else if strings.HasSuffix(v, "kb") {
		unit = KB
		subIdx = subIdx - 2
	} else if strings.HasSuffix(v, "m") {
		unit = MB
		subIdx = subIdx - 1
	} else if strings.HasSuffix(v, "mb") {
		unit = MB
		subIdx = subIdx - 2
	} else if strings.HasSuffix(v, "g") {
		unit = GB
		subIdx = subIdx - 1
	} else if strings.HasSuffix(v, "gb") {
		unit = GB
		subIdx = subIdx - 2
	}
	b, err := strconv.ParseInt(v[:subIdx], 10, 64)
	if err != nil {
		return 0, err
	}
	return int(b) * unit, nil
}

var gconf = &Config{}

func Init(file string) {
	gconf = New()
	err := gconf.Parse(file)
	if err != nil {
		panic(err)
	}
}

func Get(section string) *Section {
	return gconf.Get(section)
}

func GetKey(key string) string {
	return gconf.GetKey(key)
}

func GetKeys(key string) []string {
	return gconf.GetKeys(key)
}

func Unmarshal(v interface{}, flag ...string) error {
	f := "json"
	if len(flag) > 0 {
		f = flag[0]
	}
	return gconf.Unmarshal(v, f)
}

func UnmarshalSection(v interface{}, section string, flag ...string) error {
	f := "json"
	if len(flag) > 0 {
		f = flag[0]
	}
	return gconf.UnmarshalSection(v, section, f)
}
