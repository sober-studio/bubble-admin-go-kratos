package model

// SysPermission 权限/菜单表
type SysPermission struct {
	BaseAuthModel
	ParentID  int64  `gorm:"column:parent_id;type:bigint;default:0;comment:父权限 ID" json:"parent_id"`
	Name      string `gorm:"column:name;type:varchar(64);not null;comment:权限名称" json:"name"`
	Code      string `gorm:"column:code;type:varchar(64);not null;comment:权限编码" json:"code"`
	Type      string `gorm:"column:type;type:varchar(20);not null;comment:类型: MENU/BUTTON/API" json:"type"`
	APIPath   string `gorm:"column:api_path;type:varchar(255);comment:Kratos内部路径/API路径" json:"api_path"`
	APIMethod string `gorm:"column:api_method;type:varchar(20);default:V;comment:API方法" json:"api_method"`
	Sort      int32  `gorm:"column:sort;type:int;default:0;comment:排序" json:"sort"`
	// 前端菜单字段
	Path       string `gorm:"column:path;type:varchar(255);comment:路由路径" json:"path"`
	Component  string `gorm:"column:component;type:varchar(255);comment:组件路径" json:"component"`
	Redirect   string `gorm:"column:redirect;type:varchar(255);comment:重定向路径" json:"redirect"`
	Icon       string `gorm:"column:icon;type:varchar(64);comment:图标" json:"icon"`
	OrderNo    int32  `gorm:"column:order_no;type:int;default:0;comment:排序号" json:"order_no"`
	Hidden     bool   `gorm:"column:hidden;type:bool;default:false;comment:是否隐藏" json:"hidden"`
	KeepAlive  bool   `gorm:"column:keep_alive;type:bool;default:false;comment:是否缓存" json:"keep_alive"`
	FrameSrc   string `gorm:"column:frame_src;type:varchar(255);comment:IFrame地址" json:"frame_src"`
	FrameBlank bool   `gorm:"column:frame_blank;type:bool;default:false;comment:是否新窗口" json:"frame_blank"`
}

func (*SysPermission) TableName() string {
	return "sys_permission"
}
