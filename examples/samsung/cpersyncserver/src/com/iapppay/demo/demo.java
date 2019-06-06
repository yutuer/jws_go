package com.iapppay.demo;

import com.iapppay.sign.SignHelper;




public class demo
{
	/**
	 *类名：demo
	 *功能  服务器端签名与验签Demo
	 *版本：1.0
	 *日期：2014-06-26
	 '说明：
	 '以下代码只是为了方便商户测试而提供的样例代码，商户可以根据自己的需要，按照技术文档编写,并非一定要使用该代码。
	 '该代码仅供学习和研究爱贝云计费接口使用，只是提供一个参考。
	*/
	//
	
	public static void main(String [] argv)
	{
		
		String content = "{\"appid\":\"5000204106\",\"appuserid\":\"1:17:a73e88d9-ec6a-40f6-85b5-77c\",\"cporderid\":\"1:17:a73e88d9-ec6a-40f6-85b5-77c3f72b3275:17:101\",\"cpprivate\":\"1478487431:1:1.11.0\",\"currency\":\"RMB\",\"feetype\":0,\"money\":1.00,\"paytype\":403,\"result\":0,\"transid\":\"32041611071057135261\",\"transtime\":\"2016-11-07 10:57:26\",\"transtype\":0,\"waresid\":12}";
//		String content = "a";
		
		// 私钥
//		String priKey = "MIICdgIBADANBgkqhkiG9w0BAQEFAASCAmAwggJcAgEAAoGBAKz0WssMzD9pwfHlEPy8+NFSnsX+CeZoogRyrzAdBkILTVCukOfJeaqS07GSpVgtSk9PcFk3LqY59znddga6Kf6HA6Tpr19T3Os1U3zNeU79X/nT6haw9T4nwRDptWQdSBZmWDkY9wvA28oB3tYSULxlN/S1CEXMjmtpqNw4asHBAgMBAAECgYBzNFj+A8pROxrrC9Ai6aU7mTMVY0Ao7+1r1RCIlezDNVAMvBrdqkCWtDK6h5oHgDONXLbTVoSGSPo62x9xH7Q0NDOn8/bhuK90pVxKzCCI5v6haAg44uqbpt7fZXTNEsnveXlSeAviEKOwLkvyLeFxwTZe3NQJH8K4OqQ1KzxK+QJBANmXzpVdDZp0nAOR34BQWXHHG5aPIP3//lnYCELJUXNB2/JYTN57dv5LlE5/Ckg0Bgak764A/CX62bKhe/b+FMsCQQDLe4F2qHGy7Sa81xatm66mEkG3u88g9qRARdEvgx9SW+F1xBt2k/bU2YI31hB8IYXzL8KW9NzDfQPihBBUFn4jAkEAzbrmq/pLPlo6mHV3qE5QA2+J+hRh0UYVKsVDKkJGLH98gepS45hArbawBne/NP1bJTUVGKP9w7sl0es01hbteQJATzLO/QQq3N15Cl8dMI07uN+6PG0Y/VeCLpH+DWQXuNKSOmgN2GVW2RmfmWP0Hpxdqn2YW3EKy/vIm02TnWbzyQJAXwujUR9u9s8BZI33kw3gQ7bvWVYt8yyiYzWD2Qrnyg08tN5o+JsjW3fEDWHm70jjZIc+l/5FaZ7H5NOYpnVcpA==";
//		// 公钥
//		String pubKey = "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCs9FrLDMw/acHx5RD8vPjRUp7F/gnmaKIEcq8wHQZCC01QrpDnyXmqktOxkqVYLUpPT3BZNy6mOfc53XYGuin+hwOk6a9fU9zrNVN8zXlO/V/50+oWsPU+J8EQ6bVkHUgWZlg5GPcLwNvKAd7WElC8ZTf0tQhFzI5raajcOGrBwQIDAQAB";

		String priKey = "MIICdgIBADANBgkqhkiG9w0BAQEFAASCAmAwggJcAgEAAoGBAKoV4XnueeNAkvzsgyCAzddxGZXybbQa8ZIwAzQmDurTNzbSctHU6kkJuapsgdyD2Aq39k8nieFUYonYot3v0pqy5OzVeZDZ0mu/GiwCx9nMkKJtjlgdSaRji8ffZeUvZ/JGY+3C1pTB4v15+0ky5SqXkHoGqFgNyTXCJayAnPZDAgMBAAECgYAulGd3mRPQZLLciXkvwZad1d+H7SiWFnrp6jQ2Z+XV8ZpBbUj8pi6zafJq9eRqm8Dizpap/s4H47BIyAdyeGdYfCsEar0VaACSJcTZS5hnnzVdb5gX6pdbvun/Fi1HNZFW+XCByyNPGQuLjczg4dvFY/2bNYt6LpPLHrJsChuogQJBANV+GBdwGIqnnjOQGH+UWbxyzm3QrILxBqGltgS/8UB0obHYEZMIL6UUjDfW40J7xPXhYsNqwU7KO+Q8xs/XEfECQQDL80n2cRPSegUU/v4cbFY/g5pN01uTSCbEnMjEDL5u+Jzvf1G48peZmiw0RZ9LzWxA0jC4gTOOs0xx8TAjV1dzAkBOfq4c7/oWAMsJ6lEXl1PnFc8QUUkcW8I0bNkfpfLt3/QTj33msXvTFlr3rOqh5x/jx5qofvfUIEclA7OVd14BAkAUjGeYT95KZ4bZjbN2k6fA8HZ8ft4MIcneJ1nG/u206pGNQ8utEawaisEHZzhcf873XPYRsNrL9t6t4DoUZXlnAkEAgZa3aD/rvNPNVp64+KEFBh3XyTodrE5jmLgNBQUsV7VQNT9/qkPJrAniMys4gCeWZZCc3ayVnc/qmHO3mS+UCg==";
		// 公钥
		String pubKey = "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCn3RIAam9MMoEMtFz82DfByfMoWIcMqDtUILkzWdyXK7/oAL2bzc1BdU9BnEuI2ZoiMVWYkDoGksVxaODfbAiTjACiqsubiRT+s0TouLYiUMA3I/ncS6UEwqR0hrWIAJe0Bv8VmHcgvhu/SbMv2Uloc/eEKl4sxzMb+fMfc/80UQIDAQAB";


		// 签名
//		String sign = SignHelper.sign(content, priKey);
//		System.out.println("sign "+sign);
		// 验签
		String sign = "PCCeSBMHaiMaToOkWS/MFUqio9r80Ix1pHFW382ntQQbZ50M+HjXAGfByegcY6BhdEdStb3Yl6CkWa6Fe5wjbGCvfv2faPY/bMG05mDFd75pvU4c6FkX571rIPKEhlRlAt/MlquHMPqFFwKkspL3LmM8r7C+Ld/HUJ+9x/Bdfqs=";
		if (SignHelper.verify(content, sign, pubKey))
		{
			System.out.println("verify ok");
		}
		else
		{
			System.out.println("verify fail");
		}
		
	}
	
}
