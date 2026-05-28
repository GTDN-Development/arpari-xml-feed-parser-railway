# Feed Task: STIMA katalog produktů

## Metadata

- Supplier: STIMA
- Generator name: `stima-products`
- Output endpoint: `/feeds/stima-products.xml`
- Test generator name: `stima-products-test`
- Test output endpoint: `/feeds/stima-products-test.xml`
- Source URL: `https://www.stima.cz/userfiles/xml/ITTC_SHT_products.xml`
- Priority: první fáze
- Status: MVP implementováno, business pravidla k doplnění
- Last updated: 2026-05-28

## Cíl

Generovat Shoptet produktový XML feed ze STIMA katalogu produktů.

## Aktuální pravidla

- Parser nečte webový konfigurátor STIMA.
- Výstup je technické MVP bez SEO; cílové Shoptet kategorie, obrázky a popisy už se mapují.
- Testovací endpoint `stima-products-test` používá stejná pravidla, ale končí po prvních 2 výstupních produktech.
- Variantní parametry:
  - `KOSTRA`
  - `Sedák`
  - `Délka stolu`
  - `Rozklad`
- Produkty nad 512 variant se oříznou na prvních 512 variant v pořadí ze STIMA feedu.
- Pokud STIMA parent produkt nemá `CODE`, odvodí se z první varianty, například `ART13627-k002-l244` -> `ART13627`.
- U variantních produktů se odvozený parent kód zapisuje jako Shoptet `EXTERNAL_ID`, ne jako top-level `CODE`; Shoptet schéma jinak odmítne kombinaci parent `CODE` + `VARIANTS`.
- Kategorie se mapují na existující Shoptet kategorie z exportu `categories (1).csv`.
- Používají se jen známé cílové kategorie; nejisté STIMA kategorie jako `Katalog 2026`, `Stále skladem`, `Doprodej` nebo `Masiv dub` se zatím ignorují.
- Položka bez bezpečně určené cílové kategorie se přeskočí.
- Neřeší se rozdělení podsedáků `Skladem / Na zakázku`, dokud STIMA nepotvrdí datový zdroj.

## MVP rozsah

- Jednoduché produkty:
  - `CODE`
  - `NAME`
  - `SHORT_DESCRIPTION`
  - `DESCRIPTION`
  - `EAN`
  - `PRICE_VAT`
  - `STOCK`
- Variantní produkty:
  - parent `EXTERNAL_ID`
  - parent `NAME`
  - parent `SHORT_DESCRIPTION`
  - parent `DESCRIPTION`
  - parent `CATEGORIES`
  - parent `IMAGES`
  - variant `CODE`
  - variant `EAN`
  - variant `PRICE_VAT`
  - variant `STOCK`
  - variant `PARAMETERS`

## Aktuální ověření

- Stav kódu: implementováno v `internal/stima/products.go`, `internal/stima/categories.go` a `internal/feed/stima_products.go`.
- Registry: supplier `stima-products` a `stima-products-test` jsou dostupné přes `cmd/rebuild`.
- Lokální testy: `go test ./...` prochází.
- Reálný rebuild proti STIMA zdroji ověřen 2026-05-28.
- Reálný výstup ověřen proti Shoptet RNG schématu `products-supplier-v10.rng`.
- Poslední ověřené počty:
  - products read: 953
  - products emitted: 952
  - products skipped: 1 (`DOPRAVA` / manipulační poplatek bez cílové kategorie)
  - products trimmed over 512 variants: 20
  - variants emitted: 20581
  - variants trimmed: 3431
  - short descriptions emitted: 208
  - descriptions emitted: 207
  - image blocks: 517
  - images emitted: 1182
  - output size: přibližně 13.6 MB
- Testovací feed `stima-products-test` ověřen 2026-05-28:
  - products emitted: 2
  - variants emitted: 675
  - images emitted: 35
  - output size: přibližně 437 KB

## Otevřené otázky

- Jak ve feedu bezpečně poznat podsedáky `Skladem` vs. `Na zakázku`?
- Má katalogový import zakládat nové produkty, nebo jen připravit data pro mapování?
- Které katalogové atributy smí STIMA později přepisovat u existujících produktů?
- Jak mapovat STIMA kategorie na cílové Shoptet kategorie?
- Která katalogová pole kromě popisů, kategorií a obrázků se mají později mapovat?

## Akceptační kritéria

- `go run ./cmd/rebuild --supplier stima-products` vytvoří validní XML.
- Výstup je dostupný na `/feeds/stima-products.xml`.
- Výstup není prázdný.
- Produkty nad 512 variant neporuší Shoptet limit.
- Chyba downloadu nebo transformace nepřepíše poslední validní XML.
- `/status` ukazuje výsledek posledního běhu.

## Testy

- Unit test jednoduchého produktu.
- Unit test variantního produktu.
- Unit test odvození parent `CODE`.
- Unit test přeskočení varianty bez `CODE`.
- Unit test ořezu nad 512 variant.
- Rebuild test přes fixture-backed downloader.
- Ruční kontrola přes Shoptet XML validátor.
