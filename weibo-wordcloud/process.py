import re
import jieba
import pandas
import random
import matplotlib.pyplot as plt
from collections import Counter
from stopwordsiso import stopwords
from wordcloud import WordCloud, ImageColorGenerator

with open('data1284.txt') as f:
    data = f.read().splitlines()

with open('filter.txt') as f:
    filters = f.read().splitlines()

sub_pat = re.compile(r'<.*?>|&#.*?;|\s+|\W', re.S)

print(len(data))
print(data[:5])

print('replacing...')
data = [sub_pat.sub('', item) for item in data if item] # replace
print(len(data))
print(data[:5])

print('uniquifying...')
data = list(set(data)) # uniquify
print(len(data))
print(data[:5])

def filter_func(s):
    if not s:
        return False
    for filter_ in filters:
        if filter_ in s:
            return False
    return True
        
print('filtering...')
data = list(filter(filter_func, data)) # filter unwanted items
print(len(data))
print(data[:5])

with open('data-processed.txt', 'w') as f:
    f.write('\n'.join(data))
    
stopwords = stopwords('zh')
with open('stopwords.txt') as f:
    stopwords.update(f.read().splitlines())
assert '今日' in stopwords

all_words_to_count = []
    
for item in data:
    words = jieba.lcut(item, HMM=True)
    all_words_to_count += [w for w in words if w not in stopwords]

counter = Counter(all_words_to_count)
print(counter.most_common(100))

FONT_PATH = '/usr/share/fonts/noto-cjk/NotoSansCJK-Bold.ttc'

def color_func(*args, **kwargs):
    COLORS = ((179, 87, 43),
              (199, 81, 34),
              (200, 85, 36),
              (201, 93, 40),
              (249, 156, 52),
              (251, 196, 133),
              # (255, 251, 178),
              # (202, 226, 136),
              (203, 204, 86),
              (11, 148, 68),
              (167, 169, 172),
              (88, 87, 90),
              (77, 122, 159),
              (47, 73, 89))
    return random.choice(COLORS)

wc =  WordCloud(
    font_path=FONT_PATH,
    background_color="white",
    max_words=2000,
    # mask=back_coloring,
    # max_font_size=100,
    random_state=42,
    width=1920,
    height=1080,
    margin=2,
    color_func=color_func
)

wc.fit_words(dict(counter.most_common(2000)))
# plt.imshow(wc)
# plt.axis('off')
# plt.savefig('wordcloud.png')
# plt.show()

wc.to_file('wordcloud.png')

