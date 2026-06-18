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
  - Opakované produkty bezpečně skládá do variant podle `PRODUCT` a hodnoty z `DESCRIPTION`.
  - Mapuje cenu, sklad, dostupnost, obrázky, parametry a cílové kategorie.

- Dřevočal katalog: [drevocal.xml](https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/drevocal.xml)
  - Katalogový feed z Dřevočal B2B XML.
  - Skládá varianty matrací podle `ITEMGROUP_ID`.
  - Variantní parametry jsou `Rozměr`, `Výška` a `Potah`.
  - Mapuje produkty do kategorie `LOŽNICE > MATRACE`.
  - Přenáší dostupnost z `AVAILABILITY`.
  - Volitelný `GIFT` posílá jako doplňkový parametr `Dárek`.

- Sakypaky katalog: [sakypaky.xml](https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/sakypaky.xml)
  - Katalogový feed ze Sakypaky B2B XML.
  - Bere sedací vaky, sedací pytle, taburety, houpačky, stolky, sety, náplně a opravné sady.
  - Vynechává pelechy / psí produkty, etikety, jmenovky a položky bez bezpečně namapované kategorie.
  - Skládá varianty podle `ITEMGROUP_ID` a používá variantní parametr `Barva`.

## Autronic aktualizační feedy

- Autronic sklad: [autronic-availability.xml](https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/autronic-availability.xml)
  - Aktualizační skladový feed z Autronic availability feedu.
  - Katalogový product feed používá jen pro filtr a variantní tvar.
  - Nepřepisuje názvy, popisy, obrázky, ceny ani kategorie.

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
- Dřevočal test: [drevocal-test.xml](https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/drevocal-test.xml)
- Sakypaky test: [sakypaky-test.xml](https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/sakypaky-test.xml)

Testovací feedy jsou jen pro rychlou kontrolu importu. Nepoužívat pro ostré plnění katalogu.

## Technický feed

- Hello dummy: [hello.xml](https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/hello.xml)
  - Pouze technická kontrola, nepoužívat pro Shoptet import.
