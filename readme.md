# Go Stemmer

Go Stemmer adalah pencari akar kata bahasa Indonesia.
Go Stemmer merupakan hasil porting ke bahasa pemrograman Go dari [Pengakar](https://github.com/ivanlanin/pengakar) yang sudah dibuat oleh [Ivan Lanin](https://github.com/ivanlanin) dengan menggunakan bahasa pemrograman PHP.

## Penggunaan
Bisa dikompilasi terlebih dahulu.
```
$ go build
$ ./go-stemmer melihat
{"word":"melihat","count":1,"roots":{"lihat":{"affixes":"me-","class":"v","lemma":"lihat","prefixes":"me-"}}}
```
Atau bisa langsung dijalankan.
```
$ go run stemmer.go melihat
{"word":"melihat","count":1,"roots":{"lihat":{"affixes":"me-","class":"v","lemma":"lihat","prefixes":"me-"}}}
```
