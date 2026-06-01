# Feed Task: HON katalog

## Metadata

- Supplier: HON
- Generator name: `hon`
- Output endpoint: `/feeds/hon.xml`
- Test generator name: `hon-test`
- Test output endpoint: `/feeds/hon-test.xml`
- Source URL: `https://www.webshop.officepro-brno.cz/import/HONClientFeed/HONClientFeed.xml`
- Priority: první fáze
- Status: MVP implementováno, varianty doplněné
- Last updated: 2026-06-01

## Cíl

Generovat Shoptet XML feed z HON dodavatelského katalogu.

## Aktuální pravidla

- HON má vlastní XML strukturu.
- Parser musí vycházet z reálné struktury feedu, ne z předpokladu Zboží.cz / Heureka.
- Reálný zdroj má 517 položek.
- Testovací endpoint `hon-test` používá stejná pravidla, ale končí po prvních 5 produktech.
- Katalogová data nesmí bez mapování přepisovat citlivá data původního katalogu.
- Kategorie jsou zatím mapované široce podle `MAIN_CATEGORY` na kancelářské židle, konferenční židle, židle nebo bytové doplňky.
- Každý produkt se označuje `SUPPLIER=HON`; hodnota je určená pro interní filtrování dodavatele v Shoptet administraci.
- Zdrojové `PARAM` hodnoty se exportují jako Shoptet `INFORMATION_PARAMETERS`, tedy jako tabulkové doplňkové parametry, ne jako vybíratelné varianty.
- Variantní produkty se skládají z položek se stejným `PRODUCT` a `MAIN_CATEGORY`, pokud jde z `DESCRIPTION` bezpečně vytáhnout unikátní hodnotu varianty.
- Variantní parametr se jmenuje `Provedení`.
- Nejasné skupiny, například dlouhé rozdílné popisy bez stabilní variantní hodnoty, zůstávají jako samostatné produkty.

## MVP rozsah

- Identifikace produktu:
  - `CODE`
- Variantní produkty:
  - parent `EXTERNAL_ID`
  - variant `CODE`
  - variant `PRICE_VAT`
  - variant `STOCK`
  - variant `AVAILABILITY`
  - variant `IMAGE_REF`
  - variant `PARAMETERS`
- Bezpečná základní pole:
  - `NAME`
  - `PRICE_VAT`, pokud je ve zdroji
  - `STOCK`
  - dostupnost
  - `DESCRIPTION`
  - `IMAGES`
  - technické parametry v `INFORMATION_PARAMETERS`

## Implementace

- Stav kódu: implementováno v `internal/hon/products.go` a `internal/feed/hon_products.go`.
- Registry: supplier `hon` a `hon-test` jsou dostupné přes `cmd/rebuild`.
- Lokální testy: `go test ./...` prochází.
- Reálný rebuild ověřen 2026-06-01:
  - `hon`: 517 přečteno, 208 emitovaných `SHOPITEM`.
  - `hon`: 111 variantních produktů, 420 emitovaných variant.
  - `hon`: 0 variantních produktů s duplicitní kombinací parametrů.
  - `hon`: 208 bloků `INFORMATION_PARAMETERS`, 832 informačních parametrů.
  - `hon`: 208 bloků obrázků, 517 obrázků.
  - `hon-test`: 5 emitovaných produktů, z toho 2 variantní produkty a 9 variant.
  - Reálný výstup ověřen proti Shoptet RNG schématu `products-supplier-v10.rng`.

## Otevřené otázky

- Potvrdit, jestli `PART_NUMBER` je správný stabilní produktový kód pro párování v Shoptetu.
- Doladit cílové kategorie po ručním importním testu.
- Má HON zakládat nové produkty, nebo jen aktualizovat existující?

## Akceptační kritéria

- `go run ./cmd/rebuild --supplier hon` vytvoří validní XML.
- Výstup je dostupný na `/feeds/hon.xml`.
- `go run ./cmd/rebuild --supplier hon-test` vytvoří testovací feed s 5 produkty.
- Testovací výstup je dostupný na `/feeds/hon-test.xml`.
- Výstup nepřekročí Shoptet limity.
- Chyba downloadu nebo transformace nepřepíše poslední validní XML.
- `/status` ukazuje výsledek posledního běhu.

## Testy

- Unit test běžné HON položky.
- Unit test variantního produktu složeného podle `PRODUCT` a `DESCRIPTION`.
- Unit test nejasné skupiny, která zůstává jako samostatné produkty.
- Unit test duplicitní hodnoty varianty, která zůstává jako samostatné produkty.
- Unit test chybějícího produktu kódu.
- Unit test limitu testovacího výstupu.
- Unit test základního mapování kategorie.
- Rebuild test přes fixture-backed downloader.
- Ruční kontrola přes Shoptet XML validátor.
