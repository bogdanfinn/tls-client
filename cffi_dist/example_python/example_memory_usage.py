import ctypes
import json
import asyncio
import os, psutil

# load the tls-client shared package for your OS you are currently running your python script (i'm running on mac)
library = ctypes.cdll.LoadLibrary('./../dist/tls-client-darwin-amd64-1.2.0.dylib')

# extract the exposed request function from the shared package
request = library.request
request.argtypes = [ctypes.c_char_p]
request.restype = ctypes.c_char_p

freeMemory = library.freeMemory
freeMemory.argtypes = [ctypes.c_char_p]

destroySession = library.destroySession
destroySession.argtypes = [ctypes.c_char_p]
destroySession.restype = ctypes.c_char_p

destroyAll = library.destroyAll
destroyAll.restype = ctypes.c_char_p

async def main():
    i = 0
    while True:
        i = i + 1
        requestPayload = {
            "tlsClientIdentifier": "chrome_107",
            "followRedirects": False,
            "insecureSkipVerify": False,
            "withoutCookieJar": False,
            "isByteRequest": False,
            "forceHttp1": False,
            "withDebug": False,
            "withRandomTLSExtensionOrder": False,
            "session": i,
            "timeoutSeconds": 30,
            "proxyUrl": "",
            "headers": {
                "accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
                "user-agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36",
                "accept-encoding": "gzip, deflate, br",
                "accept-language": "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"
            },
            "headerOrder": [
                "accept",
                "user-agent",
                "accept-encoding",
                "accept-language"
            ],
            "requestUrl": "https://microsoft.com",
            "requestMethod": "GET",
            "requestBody": "",
            "requestCookies": []
        }
        request(json.dumps(requestPayload).encode('utf-8'))
        process = psutil.Process(os.getpid())
        print(process.memory_info().rss / 1024 / 1024)
        await asyncio.sleep(5)
        continue

if __name__ ==  '__main__':
    loop = asyncio.get_event_loop()
    loop.run_until_complete(main())