# Feed Task: STIMA sklad + ceny

## Metadata

- Supplier: STIMA
- Generator name: `stima-stock-price`
- Output endpoint: `/feeds/stima-stock-price.xml`
- Source URL: `https://www.stima.cz/userfiles/xml/ITTC_SHT_stock_price.xml`
- Priority: první fáze
- Status: MVP implementováno, business pravidla k doplnění
- Last updated: 2026-05-28

## Cíl

Generovat samostatný Shoptet XML feed pro aktualizaci skladů a cen STIMA produktů a variant.

## Aktuální pravidla

- Feed má aktualizovat jen skladová a cenová data.
- Nemá přepisovat katalogová pole jako názvy, popisy, obrázky, SEO nebo kategorie.
- Musí být oddělený od katalogového feedu `stima-products`.
- Chyba tohoto feedu nesmí ovlivnit `stima-products` ani `stima-stock`.
- Produkty nad 512 variant se zatím oříznou na prvních 512 variant v pořadí ze STIMA feedu.
- U variantních produktů se parent kód zapisuje jako Shoptet `EXTERNAL_ID`, ne jako top-level `CODE`; jednotlivé varianty dál nesou vlastní `CODE`.

## MVP rozsah

- Identifikace produktu nebo varianty:
  - `CODE`
  - parent `EXTERNAL_ID` u variantních produktů
- Cena:
  - `PRICE_VAT`
- Sklad:
  - `STOCK`
  - případně sklad po skladech, pokud ho zdroj obsahuje

## Aktuální ověření

- Stav kódu: implementováno v `internal/stima/updates.go` a `internal/feed/stima_updates.go`.
- Registry: supplier `stima-stock-price` je dostupný přes `cmd/rebuild`.
- Lokální testy: `go test ./...` prochází.
- Reálný rebuild proti STIMA zdroji ověřen 2026-05-28.
- Reálný výstup ověřen proti Shoptet RNG schématu `products-supplier-v10.rng`.
- Poslední ověřené počty:
  - products read: 953
  - products emitted: 953
  - products skipped: 0
  - products trimmed over 512 variants: 20
  - variants emitted: 20581
  - variants trimmed: 3431
  - output size: přibližně 12.5 MB

## Otevřené otázky

- Bude se používat samostatně vedle `stima-stock`, nebo ho později nahradí?
- Jak často se mají ceny aktualizovat?
- Je cena ze STIMA finální pro Shoptet, nebo se bude upravovat marží / pravidlem?
- Má import přepisovat cenu u existujících produktů bez ručního mapování?
- Je pro aktualizaci cen variant lepší zachovat variantní strukturu, nebo posílat varianty po jednotlivých kódech?

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
