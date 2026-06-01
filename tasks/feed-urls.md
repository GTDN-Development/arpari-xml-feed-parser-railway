# Produkční URL feedů

Base URL: `https://arpari-xml-feed-parser-railway-production.up.railway.app`

## Ostré katalogové feedy

- STIMA katalog: [stima-products.xml](https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/stima-products.xml)
  - Zakládací katalogový feed pro STIMA produkty.
  - Mapuje kategorie, obrázky, popisy, sklad a varianty.
  - Varianty jsou omezené na Shoptet limit 512 variant na produkt.

- SEGO katalog: [sego.xml](https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/sego.xml)
  - Katalogový feed ze SEGO Zboží.cz XML.
  - Skládá bezpečně rozpoznané položky do variant, ostatní nechává jako samostatné produkty.
  - Mapuje kategorie, obrázky, ceny, dostupnost a doplňkové parametry.

- Autronic katalog: [autronic-products.xml](https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/autronic-products.xml)
  - Katalogový feed z Autronic product feedu.
  - Bere nábytek a klientem vybrané bytové doplňky, sezónu vynechává.
  - Varianty skládá podle explicitních `ColorVariants`.

- HON katalog: [hon.xml](https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/hon.xml)
  - Katalogový feed z HON XML.
  - Produkty jsou jednoduché, bez variant.
  - Mapuje cenu, sklad, dostupnost, obrázky, parametry a cílové kategorie.

## STIMA aktualizační feedy

- STIMA sklad: [stima-stock.xml](https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/stima-stock.xml)
  - Aktualizuje pouze sklad.
  - Nepřepisuje názvy, popisy, obrázky ani kategorie.
  - Použít jen pokud klient nechce automaticky přepisovat ceny.

- STIMA sklad + ceny: [stima-stock-price.xml](https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/stima-stock-price.xml)
  - Aktualizuje sklad i cenu s DPH.
  - Nepřepisuje názvy, popisy, obrázky ani kategorie.
  - Doporučený update feed, pokud STIMA ceny mají být finální pro e-shop.

Poznámka: pro STIMA automatický import použít buď `stima-stock`, nebo `stima-stock-price`, ne oba najednou.

## Testovací feedy

- STIMA test: [stima-products-test.xml](https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/stima-products-test.xml)
- SEGO test: [sego-test.xml](https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/sego-test.xml)
- Autronic test: [autronic-products-test.xml](https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/autronic-products-test.xml)
- HON test: [hon-test.xml](https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/hon-test.xml)

Testovací feedy jsou jen pro rychlou kontrolu importu. Nepoužívat pro ostré plnění katalogu.

## Technický feed

- Hello dummy: [hello.xml](https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/hello.xml)
  - Pouze technická kontrola, nepoužívat pro Shoptet import.
