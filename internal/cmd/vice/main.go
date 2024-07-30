package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
	"github.com/blevesearch/vellum"
	"gopkg.in/yaml.v2"
)

type Fixture struct {
	UserAgent     string  `yaml:"user_agent"`
	Bot           *Bot    `yaml:"bot"`
	Os            *Os     `yaml:"os"`
	Client        *Client `yaml:"client"`
	Device        *Device `yaml:"device"`
	OsFamily      string  `yaml:"os_family"`
	BrowserFamily string  `yaml:"browser_family"`
}

func (f *Fixture) Setup(m *scopes) {
	if f.Os != nil {
		os := m.scope("os")
		os.get("name").add(f.Os.o.Name)
		os.get("version").add(f.Os.o.Version)
	}
	if f.Client != nil {
		b := m.scope("browser")
		b.get("name").add(f.Client.Name)
		b.get("version").add(f.Client.Version)
	}
	if f.Device != nil {
		d := m.scope("device")
		d.get("type").add(f.Device.Type)
	}
}

func (f *Fixture) Set(m *scopes, id uint64) {
	if f.Bot != nil {
		m.scope("bot").get("bot").b.SetValue(id, 1)
	}
	if f.Os != nil {
		os := m.scope("os")
		os.get("name").Set(id, f.Os.o.Name)
		os.get("version").Set(id, f.Os.o.Version)
	}
	if f.Client != nil {
		b := m.scope("browser")
		b.get("name").Set(id, f.Client.Name)
		b.get("version").Set(id, f.Client.Version)
	}
	if f.Device != nil {
		d := m.scope("device")
		d.get("type").Set(id, f.Device.Type)
	}
}

func (f *Fixture) Merge(o *Fixture) {
	if o.Bot != nil {
		f.Bot = o.Bot
	}
	if o.Os != nil {
		f.Os = o.Os
	}
	if o.Client != nil {
		f.Client = o.Client
	}
	if o.Device != nil {
		f.Device = o.Device
	}
	if o.OsFamily != "" {
		f.OsFamily = o.OsFamily
	}
	if o.BrowserFamily != "" {
		f.BrowserFamily = o.BrowserFamily
	}
}

type Bot struct {
	Name     string `yaml:"name"`
	Category string `yaml:"category"`
}

type Os struct {
	o OsImpl
}
type OsImpl struct {
	Name     string `yaml:"name"`
	Version  string `yaml:"version"`
	Platform string `yaml:"platform"`
}

var _ yaml.Unmarshaler = (*Os)(nil)

func (o *Os) UnmarshalYAML(unmarshal func(interface{}) error) error {
	unmarshal(&o.o)
	return nil
}

type Client struct {
	Name          string `yaml:"name"`
	Type          string `yaml:"type"`
	Version       string `yaml:"version"`
	Engine        string `yaml:"engine"`
	EngineVersion string `yaml:"engine_version"`
}

type Device struct {
	Type  string `yaml:"type"`
	Brand string `yaml:"brand"`
	Model string `yaml:"model"`
}

type ns map[string]*bsi

func (n ns) translate() {
	for _, v := range n {
		v.translate()
	}
}

func (b ns) get(name string) *bsi {
	if n, ok := b[name]; ok {
		return n
	}
	n := &bsi{
		name: name,
		m:    make(map[string]int),
		b:    roaring64.NewDefaultBSI(),
	}
	b[name] = n
	return n
}

type scopes struct {
	ns map[string]ns
}

func (b *scopes) write(keys []string, path string) error {
	os.MkdirAll(path, 0755)
	var buf bytes.Buffer
	for k, n := range b.ns {
		base := filepath.Join(path, k)
		os.MkdirAll(base, 0755)
		for _, m := range n {
			err := m.write(&buf, base)
			if err != nil {
				return err
			}
		}
	}
	buf.Reset()
	build, err := vellum.New(&buf, nil)
	if err != nil {
		return err
	}
	for i := range keys {
		err = build.Insert([]byte(keys[i]), uint64(i))
		if err != nil {
			return err
		}
	}
	err = build.Close()
	if err != nil {
		return err
	}
	base := filepath.Join(path, "fst")
	os.MkdirAll(base, 0755)
	file := filepath.Join(base, "fst.gz")
	return os.WriteFile(file, zip(buf.Bytes()), 0600)
}

func (b *scopes) translate() {
	for _, m := range b.ns {
		m.translate()
	}
}

func (b *scopes) scope(name string) ns {
	if n, ok := b.ns[name]; ok {
		return n
	}
	n := make(ns)
	b.ns[name] = n
	return n
}

type bsi struct {
	name string
	m    map[string]int
	keys []string
	b    *roaring64.BSI
}

func (b *bsi) write(buf *bytes.Buffer, path string) error {
	if len(b.keys) > 0 {
		data, _ := json.Marshal(b.keys)
		file := filepath.Join(path, b.name+"_translate.json.gz")
		err := os.WriteFile(file, zip(data), 0600)
		if err != nil {
			return err
		}
	}
	buf.Reset()
	b.b.WriteTo(buf)
	file := filepath.Join(path, b.name+".bsi.gz")
	return os.WriteFile(file, zip(buf.Bytes()), 0600)
}

func (b *bsi) add(name string) {
	b.m[name] = 0
}

func (b *bsi) Set(id uint64, name string) {
	b.b.SetValue(id, int64(b.m[name]))
}

func (b *bsi) translate() {
	b.keys = make([]string, 0, len(b.m))
	for k := range b.m {
		b.keys = append(b.keys, k)
	}
	slices.Sort(b.keys)
	for i := range b.keys {
		b.m[b.keys[i]] = i
	}
}

func main() {
	flag.Parse()
	root := flag.Arg(0)
	files, err := os.ReadDir(root)
	if err != nil {
		log.Fatal(err)
	}
	m := make(map[string]*Fixture)
	bm := &scopes{
		ns: make(map[string]ns),
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) != ".yml" {
			continue
		}

		o := readUA(filepath.Join(root, file.Name()))
		for _, f := range o {
			f.Setup(bm)
			g, ok := m[f.UserAgent]
			if ok {
				g.Merge(f)
				continue
			}
			m[f.UserAgent] = f
		}
	}
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, k)
	}
	sort.Strings(names)
	bm.translate()
	for i := range names {
		m[names[i]].Set(bm, uint64(i))
	}
	err = bm.write(names, "data")
	if err != nil {
		log.Fatal(err)
	}
}

func readUA(path string) (out []*Fixture) {
	f, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(f, &out)
	if err != nil {
		log.Fatal("failed to  decode ", path, err.Error())
	}
	return
}

var (
	zipBuf bytes.Buffer
	w      = gzip.NewWriter(io.Discard)
)

func zip(data []byte) []byte {
	zipBuf.Reset()
	w.Reset(&zipBuf)
	w.Write(data)
	w.Close()
	return zipBuf.Bytes()
}
