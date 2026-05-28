# Feed Task: STIMA sklad + ceny

## Metadata

- Supplier: STIMA
- Generator name: `stima-stock-price`
- Output endpoint: `/feeds/stima-stock-price.xml`
- Source URL: `https://www.stima.cz/userfiles/xml/ITTC_SHT_stock_price.xml`
- Priority: první fáze
- Status: draft, čeká na implementaci

## Cíl

Generovat samostatný Shoptet XML feed pro aktualizaci skladů a cen STIMA produktů a variant.

## Aktuální pravidla

- Feed má aktualizovat jen skladová a cenová data.
- Nemá přepisovat katalogová pole jako názvy, popisy, obrázky, SEO nebo kategorie.
- Musí být oddělený od katalogového feedu `stima-products`.
- Chyba tohoto feedu nesmí ovlivnit `stima-products` ani `stima-stock`.

## MVP rozsah

- Identifikace produktu nebo varianty:
  - `CODE`
- Cena:
  - `PRICE_VAT`
- Sklad:
  - `STOCK`
  - případně sklad po skladech, pokud ho zdroj obsahuje

## Otevřené otázky

- Bude se používat samostatně vedle `stima-stock`, nebo ho později nahradí?
- Jak často se mají ceny aktualizovat?
- Je cena ze STIMA finální pro Shoptet, nebo se bude upravovat marží / pravidlem?
- Má import přepisovat cenu u existujících produktů bez ručního mapování?

## Akceptační kritéria

- `go run ./cmd/rebuild --supplier stima-stock-price` vytvoří validní XML.
- Výstup je dostupný na `/feeds/stima-stock-price.xml`.
- Výstup obsahuje pouze bezpečná skladová a cenová pole.
- Chyba downloadu nebo transformace nepřepíše poslední validní XML.
- `/status` ukazuje výsledek posledního běhu.

## Testy

- Unit test běžné položky se skladem a cenou.
- Unit test variantní položky.
- Unit test položky bez `CODE`.
- Rebuild test přes fixture-backed downloader.
- Ruční kontrola přes Shoptet XML validátor.
