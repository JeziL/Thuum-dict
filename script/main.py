#!/usr/bin/env python3
import time
import requests
from bs4 import BeautifulSoup


class Thuum(object):
    def __init__(self, word, ipa, meanings):
        self.word = word
        self.ipa = ipa
        self.meanings = meanings

    def mdx_string(self, css_file):
        self.dragon_script = self.word.upper().replace('\'', '')
        table = {'AA': '1',
                 'AH': '4',
                 'EI': '2',
                 'EY': '9',
                 'II': '3',
                 'IR': '7',
                 'OO': '8',
                 'UU': '5',
                 'UR': '6'}
        for key, value in table.items():
            self.dragon_script = self.dragon_script.replace(key, value)
        self.body = ''
        for i, meaning in enumerate(self.meanings):
            dot_pos = meaning.find('.') + 1
            if dot_pos == 0:
                self.meanings[i - 1] = self.meanings[i - 1] + ' <i>' + meaning + '</i>'
                del self.meanings[i]
        for meaning in self.meanings:
            meaning_str = '<p><i>' + meaning + '</p>\n'
            dot_pos = meaning_str.find('.') + 1
            meaning_str = meaning_str[:dot_pos] + '</i>' + meaning_str[dot_pos:]
            self.body = self.body + meaning_str
        self.body = self.body[:-1]
        with open('template.txt', 'r') as f:
            template = f.read()
        return template.format(word=self.word,
                               css='\"' + css_file + '\"',
                               dragon_script=self.dragon_script,
                               ipa=self.ipa,
                               body=self.body)


if __name__ == '__main__':
    pages = 'ABDEFGHIJKLMNOPQRSTUVWYZ'
    thuumme = {}
    for page in pages:
        print(page + '...')
        r = requests.get('https://www.thuum.org/dictionary.php?letter=' + page)
        soup = BeautifulSoup(r.text, 'lxml')
        for div in soup.find_all('div', class_='dic-listing '):
            meanings = []
            for meaning in div.div.p.stripped_strings:
                meanings.append(meaning)
            word = div.a.string
            if word in thuumme:
                thuum = thuumme[word]
                thuum.meanings = thuum.meanings + meanings
                thuumme[word] = thuum
            else:
                thuum = Thuum(word, div.div.contents[0].strip(), meanings)
                thuumme[word] = thuum
        time.sleep(1.0)
    with open('../src/thuum.txt', 'w', newline='\r\n') as f:
        for i, (k, v) in enumerate(thuumme.items()):
            if i == len(thuumme) - 1:
                f.write(v.mdx_string(css_file='thuum.css'))
            else:
                f.write(v.mdx_string(css_file='thuum.css') + '\n')
