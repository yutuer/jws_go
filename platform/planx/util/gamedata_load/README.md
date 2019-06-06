#gamedata_load
##gamedata_load工具是什么？
    迟喻天做了三个platform/x下需要读取.data文件的项目，
    他发现每个项目都需要重新编写加载data文件的函数，很麻烦。
    gamedata_load就是解决这一问题的工具，它提供了读取单个文件的方法。
    
##我应该如何使用gamedata_load?
    gamedata_load可见两个函数：
    1.获取data文件路径
    func GetMyDataPath() string {
      	workPath, _ := os.Getwd()
      	workPath, _ = filepath.Abs(workPath)
      	// initialize default configurations
      	AppPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
      	appConfigPath := filepath.Join(AppPath, "conf")
      
      	if workPath != AppPath {
      		if utils.FileExists(appConfigPath) {
      			os.Chdir(AppPath)
      		} else {
      			appConfigPath = filepath.Join(workPath, "conf")
      		}
      	}
      	return appConfigPath
      }
    
    2.Unmarshal filepath路径下的proto.Message结构体。
    func Common_load(filepath string, v proto.Message) {
      	_errcheck := func(err error) {
      		if err != nil {
      			panic(err)
      		}
      	}
      	buffer, err := loadBin(filepath)
      	_errcheck(err)
      	err = proto.Unmarshal(buffer, v)
      	_errcheck(err)
    }
    使用事例：
    在platform/x下有一个名为gainer_beholder的工程,
    工程详情请读取gainer_beholder/README.md，
    下例是read_data.go中的初始化函数:
        func init() {
            //获取文件路径
            dataRelPath := "data"
            dataAbsPath := filepath.Join(gamedata_load.GetMyDataPath(), dataRelPath)
            //解析结构体信息
            hotpackage := &ProtobufGen.HOTPACKAGE_ARRAY{}
            gamedata_load.Common_load(filepath.Join(
            "", dataAbsPath, "hotpackage.data"), hotpackage)
            //整理data文件解析出的结构体内容，取需要的存储在全局变量中备用
            Hot_package = make(map[IDS]*ProtobufGen.HOTPACKAGE, len(hotpackage.Items))
            for _, value := range hotpackage.Items {
                key := IDS{ID: value.GetHotPackageID(), SubID: value.GetHotPackageSubID()}
                Hot_package[key] = value
            }
        }