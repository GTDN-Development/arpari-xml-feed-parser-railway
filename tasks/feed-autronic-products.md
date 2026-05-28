# Feed Task: Autronic katalog produktů

## Metadata

- Supplier: Autronic
- Generator name: `autronic-products`
- Output endpoint: `/feeds/autronic-products.xml`
- Source URL: `https://autronic.cz/feeds/product-feed.xml`
- Priority: volitelné v první fázi
- Status: draft, použít jen pokud bude potřeba katalog

## Cíl

Generovat Shoptet produktový XML feed z Autronic katalogu, pokud dostupnostní feed nebude stačit.

## Aktuální pravidla

- Autronic katalog měl přes 32 000 položek.
- Povinně importovat pouze kategorii nábytek.
- Všechny ostatní kategorie musí parser vyřadit před generováním Shoptet XML.
- Pokud i filtrovaný výstup narazí na Shoptet limit položek, bude potřeba feed dál rozdělit.
- Katalog nesmí bez mapování přepsat citlivá produktová data původního katalogu.

## MVP rozsah

- Identifikace produktu:
  - `CODE`
- Základní katalogová pole podle bezpečného zadání:
  - `NAME`
  - `EAN`
  - `PRICE_VAT`
  - `STOCK`, pokud zdroj obsahuje sklad
- Kategorie zatím jen po schválení mapování.

## Otevřené otázky

- Je katalogový feed Autronic vůbec potřeba, nebo stačí dostupnostní feed?
- Jak přesně poznat kategorii nábytek ve zdrojové struktuře?
- Kolik položek zůstane po filtru na nábytek?
- Bude potřeba výstup rozdělit na více feedů kvůli limitu Shoptetu?
- Která pole smí Autronic katalog zakládat nebo přepisovat?

## Akceptační kritéria

- `go run ./cmd/rebuild --supplier autronic-products` vytvoří validní XML.
- Výstup je dostupný na `/feeds/autronic-products.xml`.
- Do výstupu se dostanou pouze produkty z kategorie nábytek.
- Výstup nepřekročí Shoptet limit položek bez dalšího rozdělení.
- Chyba downloadu nebo transformace nepřepíše poslední validní XML.
- `/status` ukazuje výsledek posledního běhu.

## Testy

- Unit test položky z kategorie nábytek.
- Unit test přeskočení jiné kategorie.
- Unit test limitu počtu položek.
- Unit test chybějícího `CODE`.
- Rebuild test přes fixture-backed downloader.
- Ruční kontrola přes Shoptet XML validátor.
