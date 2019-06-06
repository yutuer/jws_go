# -*- coding:utf-8 -*-
from locust import HttpLocust, TaskSet, task
from auth import *
from base64 import urlsafe_b64decode, urlsafe_b64encode
import uuid
import json
headers = {"TYX-Request-Id":"taiyouxi.cn"}
def _loginDevice(l, deviceUUID, crc):
    params = {'crc':crc}
    with l.client.get(
        "/auth/v1/device/%s"%deviceUUID,
        params=params,
        name="/auth/v1/device/[device]?crc=[crc]",
        headers=headers,
        catch_response=True) as response:

        try:
            # print response.content, type(response.content)
            resp = json.loads(response.content)
            if resp['result'] != 'ok':
                response.failure("return json is not ok with %s, %s"%(
                str(resp['error']), str(resp['forclient'])))
                return
        except:
            response.failure("return is not a json")
            return

        response.success()



class NormalLoginDevice(TaskSet):
    # 设备ID匿名登录测试
    @task
    def loginAsDevice(l):
        deviceUUID = "%s@bot.com"%uuid.uuid4()
        crc = Encode64ForNet(deviceUUID)
        _loginDevice(l, deviceUUID, crc)


class LoginDeviceWrong(TaskSet):
    # 设备ID匿名登录测试，错误的crc
    @task
    def loginAsDeviceCrcWrong(l):
        deviceUUID = "%s@bot.com"%uuid.uuid4()
        crc = "none"
        _loginDevice(l, deviceUUID, crc)

def _loginReg(l, deviceUUID, name, passwd, email):
    #name应该符合英文开头，英文数字下划线
    clientuser = urlsafe_b64encode(name)
    clientpass = Encode64ForNet(passwd)
    params={
        'name': clientuser,
        'passwd': clientpass,
        'email': email,
    }
    # print name
    with l.client.get(
        "/auth/v1/user/reg/%s"%deviceUUID,
        params=params,
        name="/auth/v1/user/reg/[deviceid]?name=[name]&passwd=[passwd]&email=[email]",
        headers=headers,
        catch_response=True) as response:

        try:
            # print response.content, type(response.content)
            resp = json.loads(response.content)
            if resp['result'] != 'ok':
                response.failure("return json is not ok with %s, %s"%(
                str(resp['error']), str(resp['forclient'])))
                return
        except:
            response.failure("return is not a json")
            return

        response.success()

class RegLogin(TaskSet):
    @task
    def regAndLogin(l):
        fordid = uuid.uuid4()
        deviceUUID = "%s@reg.com"%fordid
        name = 'bot'+str(fordid).replace('-', '')
        passwd = "123456"
        email = "%s@yahoo.com"%name
        _loginReg(l, deviceUUID, name, passwd, email)


class NormalLogin(HttpLocust):
    weight = 10
    task_set = NormalLoginDevice
    min_wait=500
    max_wait=900

class WrongLogin(HttpLocust):
    weight = 0
    task_set = LoginDeviceWrong
    min_wait=5000
    max_wait=9000

class RegisterAndLogin(HttpLocust):
    weight = 10
    task_set = RegLogin
    min_wait=5000
    max_wait=9000
