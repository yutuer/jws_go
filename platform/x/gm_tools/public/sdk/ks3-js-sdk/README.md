# KS3 SDK For javascript 使用指南 #

## SDK下载地址 ##
下载地址 [https://github.com/ks3sdk/ks3-js-sdk.git](https://github.com/ks3sdk/ks3-js-sdk "https://github.com/ks3sdk/ks3-js-sdk.git")

## KS3 javascript SDK 说明##

### 概述 ###

开发者使用本 SDK 可以方便的从浏览器端上传文件至金山云存储，可使开发者忽略上传底层实现细节，而更多的关注 UI 层的展现。

### 快速入门 ###

1、通过git下载SDK到本地

`git clone https://github.com/ks3sdk/ks3-js-sdk.git`

2、SDK构成介绍及引入

- plupload.full.min.js ，建议 2.1.2 及以上版本，下载地址：[http://www.plupload.com/download/](http://www.plupload.com/download/)
- ks3jssdk.js，SDK主体文件，封装了上传功能，源码位于src目录内, 压缩版本位于dist目录内

将js文件引入到项目文件中，必须先引入`plupload.full.min.js`


```
	<script type="text/javascript" src="js/plupload.full.min.js"></script>
	<script type="text/javascript" src="js/ks3jssdk.min.js"></script>
```
3、运行

```
var ks3Options = {
    KSSAccessKeyId: "Your KSSAccessKeyId",
    policy: "Your policy",
    signature: "Your signature",
    bucket_name: "Your bucket name",
    key: '${filename}',
    uploadDomain: "http://kssws.ks-cdn.com/destination",
    autoStart: false
};
var pluploadOptions = {
    drop_element: document.body
};
var tempUpload = new ks3FileUploader(ks3Options, pluploadOptions);
document.getElementById('start-upload').onclick = function (){
        tempUpload.uploader.start()
};
```
### SDK 详细介绍 ###
#### 构造函数 ####
初始化一个 ks3FileUploader 实例：`new ks3FileUploader(ks3PostOptions, pluploadOptions);`
##### 参数 #####
- `ks3PostOptions`, 金山云存储上传需要的配置参数：

	- `KSSAccessKeyId`: AccessKey

	- `policy`: 请求中用于描述获准行为的安全策略。没有安全策略的请求被认为是匿名请求，只能访问公共可写空间。详见：[Policy、Signature构建方法](http://ks3.ksyun.com/doc/api/object/post_policy.html)
	
	- `signature`: 根据Access Key Secret和policy计算的签名信息，KS3验证该签名信息从而验证该Post请求的合法性。详见：[Policy、Signature构建方法](http://ks3.ksyun.com/doc/api/object/post_policy.html)
	
    - `bucket_name`: 上传的空间名
    
    - `key`: 被上传键值的名称。如果用户想要使用文件名作为键值，可以使用${filename} 变量。例如：如果用户想要上传文件local.jpg，需要指明specify /user/betty/${filename}，那么键值就会为/user/betty/local.jpg。
            
    - `acl`: 上传文件访问权限,有效值: private | public-read | public-read-write | authenticated-read | bucket-owner-read | bucket-owner-full-control
            
    - `uploadDomain`: 上传域名,http://destination-bucket.kss.ksyun.com 或者 http://kssws.ks-cdn.com/destination-bucket
            
    - `autoStart`: 是否在文件添加完毕后自动上传，默认为false
    
  	- `onInitCallBack`: function(){}, //上传初始化时调用的回调函数
  	
    - `onErrorCallBack`: function(){}, //发生错误时调用的回调函数
    
    - `onFilesAddedCallBack`: function(){}, //文件添加到浏览器时调用的回调函数
    
    - `onBeforeUploadCallBack`: function(){}, //文件上传之前时调用的回调函数
            
    - `onStartUploadFileCallBack`: function(){}, //文件开始上传时调用的回调函数
    
    - `onUploadProgressCallBack`: function(){}, //上传进度时调用的回调函数
    
    - `onFileUploadedCallBack`: function(){}, //文件上传完成时调用的回调函数
    
    - `onUploadCompleteCallBack`: function(){} //所有上传完成时调用的回调函数

- `pluploadOptions`, plupload上传插件需要的配置参数：
	- `runtimes`: 默认值为："html5,flash,silverlight,html4"，上传模式，上传器将会依次采用能够工作的模式运行
	- `browse_button`: 默认为'browse', 触发对话框的DOM元素自身或者其ID
	
	- `url`: 上传地址，默认为 ks3PostOptions.uploadDomain
            
    - `flash_swf_url` : 默认为'js/Moxie.swf', Flash组件的相对路径，请根据该组件在实际项目中的位置修改
    - `silverlight_xap_url` : 默认为'js/Moxie.xap', Silverlight组件的相对路径，同flash_swf_url;
	
#### 属性 ####

uploader：返回一个 plupload 插件的 Uploader 对象。 欲查看更多关于plupload的信息，参见：[http://www.plupload.com/docs/](http://www.plupload.com/docs/)

#### FAQ ####

1、 请给我一份适用于KS3 Javascript SDK policy文件示例：

答：

	{
	    "expiration": "2015-12-02T10:57:30.000Z",
	    "conditions": [{
	            "acl": "public-read"
	        }, {
	            "bucket": "test"
	        }, {
	            "key": "cloud/img/201512/ab1c145071ef03e90383ea5db4039e5d.png"
	        },
	        ["starts-with", "$name", ""],
	    ]
	}

2、 如何在初始化 ks3FileUploader 实例后更改acl, key, signature，policy等选项？

答：

	//tempUpload 为 ks3FileUploader 实例
	tempUpload.uploader.setOption("multipart_params", {
	                "key": your.key,
	                "acl": your.acl,
	                "signature" : your.signature,
	                "KSSAccessKeyId": your.KSSAccessKeyId,
	                "policy": your.policy
	            }}

3、 上传时返回403，该怎么办？

答：请做如下调试：

(1) 检查 KSSAccessKeyId 是否填写正确；

(2) 检查 policy 是否正确；

比如：在policy中定义的 acl 是 "public-read", 那么在表单项中的acl也必须是 "public-read". 

以chrome浏览器控制台为例，参见下图：

![](http://wendangimg.kssws.ks-cdn.com/javascriptsdk.png)

4、 我想在上传之前计算图片MD5, 该怎么做？

答： 下载第三方适用于javascript MD5库，比如：[https://github.com/brix/crypto-js](https://github.com/brix/crypto-js)

在文件上传之前，计算MD5, 比如：在文件添加之后：

        onFilesAddedCallBack: function(uploader, obj){
            var reader = new FileReader();
            reader.readAsBinaryString(obj[0].getSource().getSource())

            reader.onload = function(){
                console.log(CryptoJS.MD5(reader.result).toString())
            }
        }

5、为什么 onUploadProgressCallBack 中没有返回进度？

答： 在不支持HTML5的浏览器中，没有进度返回。 只能在 onErrorCallBack 回调中得到上传完成信号。

### Demo示例程序 ###
运行示例程序demo/index.html，其中包括

1. post方式上传文件到某个bucket中
`注意`：如果bucket不是公开读写的，需要先鉴权，即提供policy和signature表单域，详见main.js

2. 查看bucket中的文件对象（List Object），并转换成json格式

3. 上传图片增加水印（异步数据处理示例，只限于杭州region）
`注意`: 由于安全性考虑，由后台程序server.js计算signature，进行鉴权并请求ks3 API
server.js为一个nodeJS web服务，启动后会监听本地的3000端口

4. put上传文件

5. 分块下载大文件（支持断点续传）

6. 分块上传文件（支持断点续传）


### 许可协议 ###

	www.gnu.org/licenses/gpl-2.0.html
	
	
更多详细信息,请参阅[官方文档](http://ks3.ksyun.com/doc/api/index.html)