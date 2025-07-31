package flow

// plugin_loader.go 负责监听插件目录，动态加载 / 卸载工作流。
// 依赖 Go 原生 plugin 包及 fsnotify 文件系统事件。

import (
    "log"
    "path/filepath"
    "plugin"
    "strings"
    "sync"
    "io/fs"

    "github.com/fsnotify/fsnotify"
    "github.com/nvcnvn/adk-golang/pkg/logger"
    "go.uber.org/zap"
)

// Loader 监听插件目录并管理工作流插件的生命周期。
type Loader struct {
    Dir     string   // 插件目录
    manager *Manager // 全局 Manager
    watcher *fsnotify.Watcher
    mu      sync.Mutex
    loaded  map[string]string // flowName -> soPath
}

// NewLoader 创建 Loader。
func NewLoader(dir string, m *Manager) (*Loader, error) {
    w, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }
    if err = w.Add(dir); err != nil {
        return nil, err
    }
    l := &Loader{
        Dir:     dir,
        manager: m,
        watcher: w,
        loaded:  make(map[string]string),
    }
    // 初始加载目录中已有的 .so
    filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
        if err == nil && !d.IsDir() && strings.HasSuffix(path, ".so") {
            l.loadPlugin(path)
        }
        return nil
    })
    return l, nil
}

// Start 在 goroutine 中运行监听循环。
func (l *Loader) Start() {
    go func() {
        for ev := range l.watcher.Events {
            if ev.Op&(fsnotify.Create|fsnotify.Write) != 0 {
                if strings.HasSuffix(ev.Name, ".so") {
                    l.loadPlugin(ev.Name)
                }
            }
            if ev.Op&fsnotify.Remove != 0 {
                if strings.HasSuffix(ev.Name, ".so") {
                    l.unloadPluginByPath(ev.Name)
                }
            }
        }
    }()
}

func (l *Loader) loadPlugin(path string) {
    p, err := plugin.Open(path)
    if err != nil {
        log.Printf("[plugin_loader] 打开插件 %s 失败: %v", path, err)
        return
    }
    // 尝试向插件注入统一的 *zap.Logger
    if sym, err := p.Lookup("SetLogger"); err == nil {
        if fn, ok := sym.(func(*zap.Logger)); ok {
            fn(logger.L())
        }
    }
    sym, err := p.Lookup("Plugin")
    if err != nil {
        log.Printf("[plugin_loader] 找不到 Plugin 符号: %v", err)
        return
    }
    vptr, ok := sym.(*FlowPlugin)
    if !ok {
        log.Printf("[plugin_loader] Plugin 符号必须为 *flow.FlowPlugin 指针，实际: %T", sym)
        return
    }
    fp := *vptr
    agent, err := fp.Build()
    if err != nil {
        log.Printf("[plugin_loader] Build() 失败: %v", err)
        return
    }
    l.manager.Register(fp.Name(), agent)
    l.mu.Lock()
    l.loaded[fp.Name()] = path
    l.mu.Unlock()
    log.Printf("[plugin_loader] 已加载工作流 %s (%s)", fp.Name(), filepath.Base(path))
}

func (l *Loader) unloadPluginByPath(path string) {
    l.mu.Lock()
    defer l.mu.Unlock()
    for name, p := range l.loaded {
        if p == path {
            l.manager.Unregister(name)
            delete(l.loaded, name)
            log.Printf("[plugin_loader] 已卸载工作流 %s", name)
            return
        }
    }
}
