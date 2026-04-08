package main

// 导入所有插件包以触发 init() 函数注册
// 这些导入使用空白标识符 _ 来避免未使用的导入错误

// 资源插件
import (
	_ "mengri-flow/plugins/resource/example"
	_ "mengri-flow/plugins/resource/grpc"
	_ "mengri-flow/plugins/resource/http"
	_ "mengri-flow/plugins/resource/mysql"
	_ "mengri-flow/plugins/resource/postgres"
)

// 触发器插件
import (
	_ "mengri-flow/plugins/trigger/example_trigger"
	_ "mengri-flow/plugins/trigger/restful"
)
