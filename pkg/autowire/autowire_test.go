package autowire

import "testing"

type ITest interface {
	Test()
	Name() string
}
type TestIml1 struct {
	name   string
	tester ITest `autowired:""`
}

func (t *TestIml1) Test() {
}

func (t *TestIml1) Name() string {
	return t.name
}

// lifecycle 用于记录回调是否被触发
type lifecycle struct {
	initialized bool
	complete    bool
}

func (t *TestIml1) Initialized() {
	lifecycleFlags.initialized = true
}

func (t *TestIml1) OnComplete() {
	lifecycleFlags.complete = true
}

var lifecycleFlags lifecycle

func TestAuto(t *testing.T) {
	// 重置生命周期标记
	lifecycleFlags = lifecycle{}

	Auto(func() *TestIml1 {
		return &TestIml1{
			name: "Test1",
		}
	}, TypeName[TestIml1](), TypeName[ITest]())
	Auto(func() *int {
		i := new(int)
		*i = 99
		return i
	}, "system.port")
	var ti ITest
	var t2 *TestIml1
	var t3 = &TestIml1{
		name: "Test3",
	}
	Autowired(&ti)
	Autowired(&t2)
	Autowired(t3)
	Check()

	// Check 过程中会自动触发 OnInitialized，且在 handler 之后自动触发 OnComplete
	if !lifecycleFlags.initialized {
		t.Fatalf("Initialized should be called automatically by Check (doCheck)")
	}
	if !lifecycleFlags.complete {
		t.Fatalf("OnComplete should be called automatically after handlers in Check")
	}

	if ti != nil {
		ti.Test()
	}

	if t2 != nil {
		t2.Test()
		if t2.tester != nil {
			t2.tester.Test()
		}
	}

	if t3 != nil {
		t3.Test()
		if t3.tester != nil {
			t3.tester.Test()
		}
	}
}

// 这里不再测试 RunEvent，因为事件 handler 已并入 Check 的可选参数
