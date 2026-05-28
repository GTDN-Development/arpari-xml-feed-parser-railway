# Feed Task: HON katalog

## Metadata

- Supplier: HON
- Generator name: `hon`
- Output endpoint: `/feeds/hon.xml`
- Test generator name: `hon-test`
- Test output endpoint: `/feeds/hon-test.xml`
- Source URL: `https://www.webshop.officepro-brno.cz/import/HONClientFeed/HONClientFeed.xml`
- Priority: první fáze
- Status: MVP implementováno, čeká na ruční importní ladění
- Last updated: 2026-05-28

## Cíl

Generovat Shoptet XML feed z HON dodavatelského katalogu.

## Aktuální pravidla

- HON má vlastní XML strukturu.
- Parser musí vycházet z reálné struktury feedu, ne z předpokladu Zboží.cz / Heureka.
- Reálný zdroj má 517 položek.
- Testovací endpoint `hon-test` používá stejná pravidla, ale končí po prvních 20 produktech.
- Katalogová data nesmí bez mapování přepisovat citlivá data původního katalogu.
- Kategorie jsou zatím mapované široce podle `MAIN_CATEGORY` na kancelářské židle, konferenční židle, židle nebo bytové doplňky.

## MVP rozsah

- Identifikace produktu:
  - `CODE`
- Bezpečná základní pole:
  - `NAME`
  - `PRICE_VAT`, pokud je ve zdroji
  - `STOCK`
  - dostupnost
  - `DESCRIPTION`
  - `IMAGES`

## Implementace

- Stav kódu: implementováno v `internal/hon/products.go` a `internal/feed/hon_products.go`.
- Registry: supplier `hon` a `hon-test` jsou dostupné přes `cmd/rebuild`.
- Lokální testy: `go test ./...` prochází.
- Reálný rebuild ověřen 2026-05-28:
  - `hon`: 517 přečteno, 517 emitováno.
  - `hon-test`: 20 emitovaných produktů.
  - Výstupní XML je well-formed a publikace proběhla přes storage publisher.

## Otevřené otázky

- Potvrdit, jestli `PART_NUMBER` je správný stabilní produktový kód pro párování v Shoptetu.
- Potvrdit, zda HON položky se stejným `PRODUCT` mají zůstat samostatné produkty, nebo se mají později slučovat do variant.
- Doladit cílové kategorie po ručním importním testu.
- Má HON zakládat nové produkty, nebo jen aktualizovat existující?

## Akceptační kritéria

- `go run ./cmd/rebuild --supplier hon` vytvoří validní XML.
- Výstup je dostupný na `/feeds/hon.xml`.
- `go run ./cmd/rebuild --supplier hon-test` vytvoří testovací feed s 20 produkty.
- Testovací výstup je dostupný na `/feeds/hon-test.xml`.
- Výstup nepřekročí Shoptet limity.
- Chyba downloadu nebo transformace nepřepíše poslední validní XML.
- `/status` ukazuje výsledek posledního běhu.

## Testy

- Unit test běžné HON položky.
- Unit test chybějícího produktu kódu.
- Unit test limitu testovacího výstupu.
- Unit test základního mapování kategorie.
- Rebuild test přes fixture-backed downloader.
- Ruční kontrola přes Shoptet XML validátor.
