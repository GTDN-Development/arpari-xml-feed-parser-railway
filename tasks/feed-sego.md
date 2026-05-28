# Feed Task: SEGO katalog

## Metadata

- Supplier: SEGO
- Generator name: `sego`
- Output endpoint: `/feeds/sego.xml`
- Test generator name: `sego-test`
- Test output endpoint: `/feeds/sego-test.xml`
- Source URL: `https://segocz.cz/src/Frontend/Files/Feeds/Catalog/zbozi_123456.xml`
- Priority: první fáze
- Status: MVP implementováno, čeká na ruční importní ladění
- Last updated: 2026-05-28

## Cíl

Generovat Shoptet XML feed ze SEGO katalogového feedu.

## Aktuální pravidla

- Zdroj je ve stylu Zboží.cz.
- Transformace do Shoptet XML struktury je implementovaná jako katalogový MVP.
- Reálný zdroj má 144 položek.
- Testovací endpoint `sego-test` používá stejná pravidla, ale končí po prvních 20 produktech.
- Katalogová data nesmí bez mapování přepisovat citlivá data původního katalogu.
- Kategorie jsou zatím mapované široce do `KANCELÁŘSKÉ ŽIDLE A KŘESLA`, případně `ŽIDLE > KONFERENČNÍ ŽIDLE` podle názvu produktu.

## MVP rozsah

- Identifikace produktu:
  - `CODE`
- Bezpečná základní pole:
  - `NAME`
  - `EAN`, pokud je ve zdroji
  - `PRICE_VAT`, pokud je ve zdroji
  - dostupnost z `DELIVERY_DATE`
  - `DESCRIPTION`
  - `IMAGES`

## Implementace

- Stav kódu: implementováno v `internal/sego/products.go` a `internal/feed/sego_products.go`.
- Registry: supplier `sego` a `sego-test` jsou dostupné přes `cmd/rebuild`.
- Lokální testy: `go test ./...` prochází.
- Reálný rebuild ověřen 2026-05-28:
  - `sego`: 144 přečteno, 144 emitováno.
  - `sego-test`: 20 emitovaných produktů.
  - Výstupní XML je well-formed a publikace proběhla přes storage publisher.

## Otevřené otázky

- Potvrdit, jestli `ITEM_ID` je správný stabilní produktový kód pro párování v Shoptetu.
- Potvrdit, zda SEGO produkty mají zůstat jako jednoduché produkty, nebo se mají později slučovat do variant.
- Doladit cílové kategorie po ručním importním testu.
- Má SEGO zakládat nové produkty, nebo jen aktualizovat existující?

## Akceptační kritéria

- `go run ./cmd/rebuild --supplier sego` vytvoří validní XML.
- Výstup je dostupný na `/feeds/sego.xml`.
- `go run ./cmd/rebuild --supplier sego-test` vytvoří testovací feed s 20 produkty.
- Testovací výstup je dostupný na `/feeds/sego-test.xml`.
- Výstup nepřekročí Shoptet limity.
- Chyba downloadu nebo transformace nepřepíše poslední validní XML.
- `/status` ukazuje výsledek posledního běhu.

## Testy

- Unit test běžné SEGO položky.
- Unit test chybějícího produktu kódu.
- Unit test limitu testovacího výstupu.
- Unit test základního mapování kategorie.
- Rebuild test přes fixture-backed downloader.
- Ruční kontrola přes Shoptet XML validátor.
