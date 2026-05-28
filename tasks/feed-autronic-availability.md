# Feed Task: Autronic dostupnost

## Metadata

- Supplier: Autronic
- Generator name: `autronic-availability`
- Output endpoint: `/feeds/autronic-availability.xml`
- Source URL: `https://autronic.cz/feeds/availability-feed.xml`
- Priority: první fáze
- Status: draft, čeká na implementaci

## Cíl

Generovat Shoptet XML feed pro aktualizaci dostupnosti / skladu Autronic produktů.

## Aktuální pravidla

- Zdroj je dostupnostní feed, ne kompletní produktový katalog.
- Parser musí používat `GET`; `HEAD` u tohoto zdroje vrací 404.
- Z Autronicu se mají brát pouze produkty z kategorie nábytek.
- Produkty z ostatních kategorií musí parser přeskočit.
- Feed nemá přepisovat katalogová pole, pokud to nebude výslovně schválené.

## MVP rozsah

- Identifikace produktu:
  - `CODE`
- Dostupnost / sklad:
  - `STOCK`
  - případně `AVAILABILITY`, pokud to bude vhodnější pro Shoptet import

## Otevřené otázky

- Obsahuje dostupnostní feed kategorii, podle které lze poznat nábytek?
- Pokud dostupnostní feed kategorii neobsahuje, spojíme ho s Autronic katalogem, nebo s ručním mapováním?
- Jak přesně mapovat Autronic dostupnost na Shoptet sklad / dostupnost?
- Jaký identifikátor odpovídá produktům v cílovém katalogu?

## Akceptační kritéria

- `go run ./cmd/rebuild --supplier autronic-availability` vytvoří validní XML.
- Výstup je dostupný na `/feeds/autronic-availability.xml`.
- Parser používá `GET`, ne `HEAD`.
- Do výstupu se dostanou pouze produkty z kategorie nábytek nebo položky povolené mapováním.
- Chyba downloadu nebo transformace nepřepíše poslední validní XML.
- `/status` ukazuje výsledek posledního běhu.

## Testy

- Unit test položky z kategorie nábytek.
- Unit test přeskočení jiné kategorie.
- Unit test zdroje bez kategorie.
- Unit test chybějícího `CODE`.
- Rebuild test přes fixture-backed downloader.
- Ruční kontrola přes Shoptet XML validátor.
