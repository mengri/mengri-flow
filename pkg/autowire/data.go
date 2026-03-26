package autowire

import (
	"sync"
)

var (
	beans     = make(map[string]*bean) //  可用用注入的实体
	factories = make(map[string]*factoryItem)
	lock      sync.RWMutex
	requireBy = make(map[string][]*requireItem)
	isChecked bool // 是否已经检查过, 如果检查过,不能再执行
)
