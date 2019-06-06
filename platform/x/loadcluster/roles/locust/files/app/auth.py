# from locust import HttpLocust, TaskSet
import uuid
import md5
from string import maketrans
from base64 import urlsafe_b64decode, urlsafe_b64encode

urlBase64 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
defaultEncodePwd = "f1zJ7AerIatj_gncV0y26ZwqTS4vF3okHxRUQY8GENBlK5O9PpuCLhWDMsXb-dmi"
defaultSalt = "34rfdsf3rdkunj;l9"

base64fixTable = maketrans(defaultEncodePwd, urlBase64);
base64fixTableReverse = maketrans(urlBase64, defaultEncodePwd);

def Decode64FromNet(raw):
    return urlsafe_b64decode(raw.translate(base64fixTable))

def Encode64ForNet(raw):
    return urlsafe_b64encode(raw).translate(base64fixTableReverse)

def PasswordRawForClient(rawPasswd):
    m = md5.new()
    m.update(rawPasswd)
    return m.hexdigest()

# print PasswordRawForClient("abcdef")
# print Decode64FromNet("3qgYF8sxvw6=")
# print Encode64ForNet("username")


# def login(l):
#     l.client.post("/login", {"username":"ellen_key", "password":"education"})
#
# def index(l):
#     l.client.get("/auth/v1/device/test")
#
# def profile(l):
#     l.client.get("/profile")
#
# class UserBehavior(TaskSet):
#     # tasks = {index:2, profile:1}
#     tasks = {index:2}
#
#     def on_start(self):
#         pass
#         # login(self)
#
# class WebsiteUser(HttpLocust):
#     task_set = UserBehavior
#     min_wait=5000
#     max_wait=9000
