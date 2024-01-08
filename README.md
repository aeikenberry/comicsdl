# Comicsdl
CLI for downloading comics

### Install
`go install github.com/aeikenberry/comicsdl`

### Usaage
```
❯ comicsdl --help
Usage of comicsdl:
  -comic string
    	What comic are we shooting for? (default "Swamp Thing")
  -dest string
    	Where do you want to save the comic? (default "downloads/")
```

```
❯ comicsdl -comic "Savage Dragon" -dest dls
Searching : Savage Dragon
0: Savage Dragon #266 (2023)
1: Savage Dragon – Teenage Mutant Ninja Turtles Crossover (1993)
2: Teenage Mutant Ninja Turtles – Savage Dragon Crossover (1995)
3: Savage Dragon #265 (2023)
4: Savage Dragon #264 (2023)
5: Savage Dragon #263 (2023)
6: Savage Dragon Vol. 1 #1 – 3 + TPB (1992+2004)
7: Savage Dragon Archives Vol. 1 – 10 (2006-2021)
8: Savage Dragon #262 (2022)
9: Savage Dragon #261 (2022)
10: Savage Dragon #260 (2021)
11: Savage Avengers Vol. 3 – Enter The Dragon (TPB) (2021)
make selection: [0]
10
You chose: Savage Dragon #260 (2021)
Downloading: Savage Dragon 260 (2021) (digital) (The Seeker-Empire).cbr
Done! Saved: dls/Savage Dragon 260 (2021) (digital) (The Seeker-Empire).cbr
```
