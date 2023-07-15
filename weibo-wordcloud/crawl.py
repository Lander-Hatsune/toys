import re
import time
import requests

pattern = re.compile(
    r'feed_list_content.*?>(.*?)</p>',
    re.S, 
)

URL = 
COOKIES = 
USER_AGENT = 

headers = {
    'cookie': COOKIES,
    'user-agent': USER_AGENT,
}

params = {
    'q': '张江',
    'page': 1,
}

data = []

for page in range(1, 10000):
    
    params['page'] = page
    
    resp = requests.get(URL, params=params, headers=headers)
    assert resp.status_code == 200, "resp error"
    slic = pattern.findall(resp.text)
    data += slic
    print(f'page {page} done, get {len(slic)} items, all {len(data)} items')

    if page % 100 == 0:
        with open('data.txt', 'a+') as f:
            f.write('\n'.join(data))

    time.sleep(1)
