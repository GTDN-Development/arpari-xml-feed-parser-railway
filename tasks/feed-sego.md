# Feed Task: SEGO katalog

## Metadata

- Supplier: SEGO
- Generator name: `sego`
- Output endpoint: `/feeds/sego.xml`
- Source URL: `https://segocz.cz/src/Frontend/Files/Feeds/Catalog/zbozi_123456.xml`
- Priority: první fáze
- Status: draft, čeká na implementaci

## Cíl

Generovat Shoptet XML feed ze SEGO katalogového feedu.

## Aktuální pravidla

- Zdroj je pravděpodobně ve stylu Zboží.cz / Heureka.
- Bude nutná transformace do Shoptet XML struktury.
- Katalogová data nesmí bez mapování přepisovat citlivá data původního katalogu.
- Kategorie bude potřeba řešit přes společné mapování.

## MVP rozsah

- Identifikace produktu:
  - `CODE`
- Bezpečná základní pole:
  - `NAME`
  - `EAN`, pokud je ve zdroji
  - `PRICE_VAT`, pokud je ve zdroji
  - `STOCK` nebo dostupnost, pokud je ve zdroji

## Otevřené otázky

- Jaká je přesná XML struktura SEGO feedu?
- Které pole je stabilní produktový kód?
- Obsahuje feed varianty?
- Jak mapovat SEGO kategorie na Shoptet?
- Má SEGO zakládat nové produkty, nebo jen aktualizovat existující?

## Akceptační kritéria

- `go run ./cmd/rebuild --supplier sego` vytvoří validní XML.
- Výstup je dostupný na `/feeds/sego.xml`.
- Výstup nepřekročí Shoptet limity.
- Chyba downloadu nebo transformace nepřepíše poslední validní XML.
- `/status` ukazuje výsledek posledního běhu.

## Testy

- Unit test běžné SEGO položky.
- Unit test chybějícího produktu kódu.
- Unit test variant, pokud je zdroj obsahuje.
- Unit test kategorie po doplnění mapování.
- Rebuild test přes fixture-backed downloader.
- Ruční kontrola přes Shoptet XML validátor.
