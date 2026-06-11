# Obecné zadání: ARPARI -> Shoptet feed parser / transformer

## Shrnutí

Potřebujeme vytvořit vlastní nástroj pro zpracování produktových, skladových a cenových feedů při migraci e-shopu ARPARI z PrestaShopu na Shoptet.

Klient nechce použít NapojSe. Řešení tedy bude vlastní mezivrstva:

```text
vstupní feedy / export z původního e-shopu
-> parser / transformer
-> validní Shoptet XML
-> automatické importy v Shoptetu
```

Konkrétní stack, jazyk, framework a architektura se rozhodnou později. Důležité je, aby řešení bylo konfigurovatelné, bezpečné pro automatický provoz, validovalo výstup a umožnilo postupně přidávat další dodavatele.

## Kontext migrace

Nejdříve se do Shoptetu importují původní produkty z aktuálního PrestaShopu. Tento import připravuje kolega přímo z původní SQL databáze.

K dispozici bude:

- export původních produktů,
- export kategorií,
- vazby produktů na kategorie,
- případně další data z původního e-shopu.

Původní katalog je primární zdroj pro:

- názvy produktů,
- popisy,
- URL,
- SEO,
- původní kategorie,
- ruční úpravy,
- existující sortiment.

Dodavatelské feedy se budou napojovat až následně. Nesmí bez kontroly přepisovat původní katalog, aby nevznikaly duplicity nebo rozbité URL a SEO.

## Hlavní cíl

Nástroj má připravit validní XML feedy pro Shoptet tak, aby bylo možné:

- založit nové produkty z dodavatelských feedů,
- aktualizovat existující produkty,
- aktualizovat sklad,
- aktualizovat ceny,
- zachovat původní katalogová data tam, kde jsou důležitá,
- zabránit duplicitám,
- mapovat dodavatelské kategorie na cílové kategorie v Shoptetu,
- převádět varianty a příplatky do nativního Shoptet modelu.

## Zdroje dat

### Původní e-shop / PrestaShop

Import původních produktů zatím není finálně specifikovaný. Doplní se později.

Očekávání:

```text
originální produkty -> základní katalog v Shoptetu
```

Dodavatelské feedy pak mají sloužit jako doplnění nebo aktualizace, ne jako bezhlavé přepsání katalogu.

### Fázování dodavatelských feedů

První fáze se bude věnovat těmto zdrojům:

- STIMA katalog,
- STIMA sklad,
- STIMA sklad + ceny,
- Autronic dostupnost,
- případně Autronic katalog,
- SEGO,
- HON.

Do druhé fáze se odkládají:

- Sakypaky s.r.o.: `https://www.sakypaky.cz/export/b2b_partners_cs.xml`,
- Dřevočal s.r.o.: `https://www.matrace-drevocal.cz/feed-b2b.xml`.

### STIMA katalog

```text
https://www.stima.cz/userfiles/xml/ITTC_SHT_products.xml
```

Obsahuje kompletní katalog STIMA včetně variant, EAN kódů, cen, skladů a kategorií.

Feed je blízký Shoptet XML a validuje, ale pravděpodobně ho i tak budeme transformovat kvůli napojení na původní katalog a vlastní pravidla.

### STIMA sklad

```text
https://www.stima.cz/userfiles/xml/ITTC_SHT_stock.xml
```

Menší feed pro pravidelnou aktualizaci skladů.

### STIMA sklad + ceny

```text
https://www.stima.cz/userfiles/xml/ITTC_SHT_stock_price.xml
```

Feed pro aktualizaci skladů a cen.

### Autronic dostupnost

```text
https://autronic.cz/feeds/availability-feed.xml
```

Toto je dostupnostní feed, ne kompletní produktový katalog.

Z Autronicu se budou brát pouze produkty z kategorie nábytek. Produkty z ostatních kategorií nejsou součástí zadání a parser je musí přeskočit.

Obsahuje hlavně:

- `ProductCode`,
- `AvailabilityStatus`,
- `StockAvailabilityTotal Quantity`,
- `StockAvailability / Stock Name / Quantity`,
- `OnTheWayAvailability`.

Poznámka: `HEAD` request vrací 404, ale běžný `GET` vrací XML. Parser musí používat `GET`.

Dříve byl dohledaný i katalogový feed:

```text
https://autronic.cz/feeds/product-feed.xml
```

Ten měl přes 32 000 položek. Pokud se bude používat, je nutné filtrovat nebo rozdělit výstup, protože Shoptet má limit přibližně 20 000 položek na jeden feed.

Povinný filtr pro Autronic katalog: importovat pouze kategorii nábytek. Všechny ostatní kategorie musí být vyřazeny ještě před generováním Shoptet XML.

### SEGO

```text
https://segocz.cz/src/Frontend/Files/Feeds/Catalog/zbozi_123456.xml
```

Feed ve stylu Zboží.cz, ne Shoptet XML. Nutná transformace.

### HON

```text
https://www.webshop.officepro-brno.cz/import/HONClientFeed/HONClientFeed.xml
```

Vlastní XML struktura. Nutná transformace.

### Sakypaky

```text
https://www.sakypaky.cz/export/b2b_partners_cs.xml
```

Feed podobný Heureka/Zboží stylu. Nutná transformace. Tento feed je odložený do druhé fáze a není součástí prvního MVP.

### Dřevočal

```text
https://www.matrace-drevocal.cz/feed-b2b.xml
```

B2B XML feed podle dokumentace `reference/drevocal/drevocal-b2b-feed-dokumentace-2026-05.pdf`.
Jedna položka ve zdroji odpovídá jedné variantě matrace. Varianty jedné matrace se
sdružují přes `ITEMGROUP_ID`, kód varianty je `ITEM_ID` a variantní parametry jsou
`Rozměr`, `Výška` a `Potah`. Feed neobsahuje sklad ani dostupnost. Tento feed je
odložený do druhé fáze a není součástí prvního MVP.

## Shoptet výstupy

Výsledkem první fáze mají být veřejně dostupné XML feedy pro Shoptet automatické importy, například:

```text
/feeds/stima-products.xml
/feeds/stima-stock.xml
/feeds/stima-stock-price.xml
/feeds/sego.xml
/feeds/hon.xml
/feeds/autronic-availability.xml
```

Do druhé fáze se odkládají:

```text
/feeds/sakypaky.xml
/feeds/drevocal.xml
/feeds/drevocal-test.xml
```

Nástroj by měl mít i stavový výstup, kde bude vidět:

- kdy proběhl poslední běh,
- zda doběhl úspěšně,
- kolik položek bylo zpracováno,
- kolik položek bylo přeskočeno,
- případné chyby.

## Bezpečnost výstupu

Důležité pravidlo:

```text
Nikdy nepřepsat poslední funkční XML rozbitým feedem.
```

Doporučený princip:

1. stáhnout zdrojový feed,
2. převést data,
3. aplikovat mapování a pravidla,
4. vygenerovat dočasné XML,
5. zvalidovat výstup,
6. teprve potom nahradit veřejný XML feed.

Pokud transformace nebo validace selže, Shoptet má dál dostávat poslední validní XML.

## Párování produktů

Klíčové je správné párování produktů a variant.

Základní pravidlo:

```text
Kód produktu / varianty je hlavní identifikátor.
```

Bude potřeba mapování mezi:

```text
dodavatelský kód -> Shoptet kód / původní kód z PrestaShopu
```

U existujících produktů z původního e-shopu nepřepisovat automaticky citlivá data, pokud to nebude výslovně povoleno:

- název,
- popis,
- URL,
- SEO,
- hlavní kategorii,
- ručně upravené obrázky.

Typicky aktualizovat:

- sklad,
- cenu,
- dostupnost,
- EAN,
- variantní data,
- případně vybrané parametry.

## Varianty a příplatky

Nativní model Shoptetu:

```text
technické/skladové rozdíly = varianty
placené zákaznické volby = příplatkové parametry
```

Příklad STIMA:

```text
KOSTRA + Sedák = variantní parametry
```

Původní produkty z e-shopu mohou mít volby typu látky, potahy, kůže, vzory.

Bez custom konfigurátoru se to v Shoptetu řeší jako jeden příplatkový parametr, například:

```text
Potah
```

Hodnoty:

```text
Látky 1 | Bombay 34 = +0 Kč
Látky 2 | Valencia 10 = +545 Kč
Látky 3 | Velvet 60 = +1210 Kč
Látky 4 / Koženky 4 | Silver 22 = +3025 Kč
Kůže | Černá = +7865 Kč
```

Shoptet nativně neumí závislé výběry typu:

```text
nejdřív skupina látky -> potom jen vzory z dané skupiny
```

Bez custom frontendu tedy půjde o jeden delší seznam hodnot.

## Limity k hlídání

Shoptet limity, které je potřeba respektovat:

- max přibližně 20 000 položek na jeden feed,
- max 3 variantní parametry na produkt,
- max 512 variant na produkt,
- max přibližně 6 příplatkových parametrů na produkt,
- max přibližně 128 hodnot jednoho příplatkového parametru.

Známý problém:

```text
STIMA produkt Židle KR18 má ve feedu přibližně 540 variant.
```

To může překročit limit Shoptetu 512 variant na produkt a bude potřeba to řešit.

Autronic katalogový feed měl přes 32 000 položek. Pokud se použije, bude nutné ho filtrovat na kategorii nábytek; případné další dělení výstupu se řeší až pokud i filtrovaný výstup narazí na limit Shoptetu.

## Importní strategie

Správný postup:

1. import původního katalogu z PrestaShopu,
2. kontrola kódů, kategorií, URL a variant,
3. příprava mapování původní katalog <-> dodavatelské feedy,
4. napojení STIMA,
5. postupné napojení dalších dodavatelů, přičemž Sakypaky a Dřevočal až ve druhé fázi,
6. první běhy importů nastavit konzervativně.

Konzervativní nastavení prvních importů:

- nemazat produkty,
- neskrývat automaticky produkty chybějící ve feedu,
- nepřepisovat názvy, popisy ani SEO bez kontroly.

## Oficiální podklady Shoptet

- [Shoptet XML specifikace](https://developers.shoptet.com/shoptet-tools/shoptet-xml-specification/)
- [Automatické importy produktů](https://podpora.shoptet.cz/automaticke-importy-produktu/)
- [Šablony variant](https://podpora.shoptet.cz/sablony-variant/)
- [Příplatkové parametry](https://podpora.shoptet.cz/priplatkove-parametry/)

## Hosting a provoz

Řešení bude hostované na platformě Railway.

Požadavky:

- veřejně dostupné HTTPS URL pro výstupní XML feedy,
- pravidelné spouštění transformací přes plánovaný job nebo cron,
- možnost ručně spustit rebuild feedů,
- stavový endpoint nebo jednoduchý status výstup,
- persistentní uložení posledních validních XML výstupů,
- logování běhů a chyb v Railway logs,
- konfigurace přes environment variables a konfigurační soubory,
- nespoléhat na lokální ephemeral filesystem pro jedinou kopii feedů, pokud by mohl být při redeployi smazán.

Doporučený princip provozu:

1. cron/job stáhne zdrojové feedy,
2. parser vygeneruje nové XML do dočasného výstupu,
3. XML se zvaliduje,
4. pokud validace projde, uloží se jako aktuální veřejný feed,
5. pokud validace selže, veřejně zůstává poslední validní feed.

## Požadavky na implementaci

Řešení má být:

- konfigurovatelné,
- validovat výstup,
- logovat chyby,
- bezpečně držet poslední funkční feed,
- umožnit postupně přidávat další dodavatele,
- nevyžadovat ruční zásahy při každé běžné aktualizaci cen nebo skladů.

## Otevřené otázky před návrhem stacku

- Jaký bude finální formát exportu původního PrestaShop katalogu?
- Kde bude uložená persistentní kopie posledních validních feedů?
- Jak bude vypadat mapování dodavatelských kódů na Shoptet / původní kódy?
- Jak bude vypadat mapování dodavatelských kategorií na cílové kategorie?
- Budou feedy validované jen strukturálně, nebo i proti XSD / pravidlům Shoptetu?
- Jak se vyřeší produkty překračující limity variant nebo příplatkových parametrů?
- Které atributy smí dodavatelský feed přepisovat u existujících produktů?
- Jaký bude minimální rozsah prvního MVP bez feedů Sakypaky a Dřevočal?
