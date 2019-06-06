/**
 * Created by libingbing on 2017/5/27.
 */

var ks3Options = {
    KSSAccessKeyId: "Your KSSAccessKeyId",
    policy: "Your policy",
    signature: "Your signature",
    bucket_name: "Your bucket name",
    key: '${filename}',
    uploadDomain: "http://kss.ksyun.com/destination", //杭州region
    autoStart: false
};
var pluploadOptions = {
    drop_element: document.body
};
var tempUpload = new ks3FileUploader(ks3Options, pluploadOptions);

