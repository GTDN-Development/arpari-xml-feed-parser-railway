# Feed Task: STIMA katalog produktů

## Metadata

- Supplier: STIMA
- Generator name: `stima-products`
- Output endpoint: `/feeds/stima-products.xml`
- Source URL: `https://www.stima.cz/userfiles/xml/ITTC_SHT_products.xml`
- Priority: první fáze
- Status: MVP implementováno, business pravidla k doplnění

## Cíl

Generovat Shoptet produktový XML feed ze STIMA katalogu produktů.

## Aktuální pravidla

- Parser nečte webový konfigurátor STIMA.
- Výstup je technické MVP bez kategorií, obrázků, popisů a SEO.
- Variantní parametry:
  - `KOSTRA`
  - `Sedák`
  - `Délka stolu`
  - `Rozklad`
- Produkty nad 512 variant se oříznou na prvních 512 variant v pořadí ze STIMA feedu.
- Pokud STIMA parent produkt nemá `CODE`, odvodí se z první varianty, například `ART13627-k002-l244` -> `ART13627`.
- Neřeší se rozdělení podsedáků `Skladem / Na zakázku`, dokud STIMA nepotvrdí datový zdroj.

## MVP rozsah

- Jednoduché produkty:
  - `CODE`
  - `NAME`
  - `EAN`
  - `PRICE_VAT`
  - `STOCK`
- Variantní produkty:
  - parent `CODE`
  - parent `NAME`
  - variant `CODE`
  - variant `EAN`
  - variant `PRICE_VAT`
  - variant `STOCK`
  - variant `PARAMETERS`

## Otevřené otázky

- Jak ve feedu bezpečně poznat podsedáky `Skladem` vs. `Na zakázku`?
- Má katalogový import zakládat nové produkty, nebo jen připravit data pro mapování?
- Které katalogové atributy smí STIMA později přepisovat u existujících produktů?
- Jak mapovat STIMA kategorie na cílové Shoptet kategorie?
- Jak doplnit obrázky a popisy bez rizika přepsání původního katalogu?

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
