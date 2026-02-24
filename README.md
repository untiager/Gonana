# Gonana - Vérificateur de Style Epitech

[![Tests](https://github.com/untiager/Gonana/actions/workflows/test.yml/badge.svg)](https://github.com/untiager/Gonana/actions/workflows/test.yml)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Coverage](https://img.shields.io/badge/coverage-89.3%25-brightgreen)](https://github.com/untiager/Gonana)

Gonana est un outil en ligne de commande développé en Go pour analyser automatiquement la conformité des fichiers C (.c) et headers (.h) avec la norme de style Epitech.

## Fonctionnalités

### Vérifications de Base (Niveau 1)
-  Taille maximale d'une ligne (80 caractères)
-  Aucune ligne vide en début/fin de fichier
-  Aucune ligne vide consécutive
-  Indentation en TAB uniquement
-  Une seule variable déclarée par ligne
-  Déclarations de variables en début de fonction uniquement
-  Nom de fichier en snake_case
-  Nom de fonction en snake_case
-  Nom de macro en SCREAMING_SNAKE_CASE
-  Fonction de 25 lignes maximum
-  Fichier de 3 fonctions maximum (hors main)

### Vérifications Avancées (Niveau 2)
-  Format de commentaires correct (/* */ uniquement)
-  Commentaire de fonction obligatoire
-  Pas de déclaration globale non const
-  Maximum 4 paramètres par fonction
-  Pas de déclaration dans les boucles for

### Fonctionnalités Complémentaires
-  Rapport détaillé dans le terminal
-  Score global de conformité
-  Sortie JSON pour automatisation
-  Interface colorée et intuitive

## Installation

### Prérequis
- Go 1.21 ou supérieur

### Compilation
```bash
```

##  Utilisation

### Syntaxe de base
```bash
Gonana [options] <fichier_ou_dossier>
```

### Options disponibles
- `-path` : Chemin du fichier ou dossier à analyser
- `-verbose` : Affichage détaillé des violations
- `-json` : Sortie au format JSON
- `-silent` : Mode silencieux (code de retour uniquement)
- `-level` : Niveau de vérification (1=base, 2=avancé)

### Exemples d'utilisation

```bash
# Analyser un fichier
Gonana mon_fichier.c

# Analyser un dossier avec sortie détaillée
Gonana -verbose src/

# Générer un rapport JSON
Gonana -json -level 2 projet/

# Mode silencieux pour scripts
Gonana -silent fichier.c
echo $?  # 0 = succès, 1 = violations détectées
```

##  Format de Sortie

### Sortie Standard
```
╔══════════════════════════════════════════════════════════════════════════════╗
║                            Gonana - RAPPORT D'ANALYSE                        ║
╚══════════════════════════════════════════════════════════════════════════════╝

 RÉSUMÉ GLOBAL
   • Fichiers analysés: 3
   • Lignes de code: 127
   • Violations totales: 5
   • Fichiers propres: 1/3
   • Propreté: 33.3% [████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░] 33.3%

 utils.c (95.2% - 42 lignes)
 main.c (78.5% - 65 lignes - 3 violations)
 parser.c (82.1% - 20 lignes - 2 violations)

╔══════════════════════════════════════════════════════════════════════════════╗
║                             SCORE GLOBAL: 85.3%                              ║
║       [██████████████████████████████████████████████░░░░░░░░] 85.3%         ║
║                  TRÈS BIEN! Quelques petits détails à corriger.              ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

### Sortie JSON
```json
{
  "files": [
    {
      "filename": "main.c",
      "violations": [
        {
          "rule": "C-L1",
          "message": "Ligne trop longue",
          "line": 15,
          "severity": "major",
          "description": "La ligne contient plus de 80 caractères"
        }
      ],
      "score": 78.5,
      "line_count": 65
    }
  ],
  "total_score": 85.3,
  "total_files": 3,
  "total_lines": 127,
  "total_violations": 5,
  "clean_files": 1
}
```

## Architecture du Projet

```
Gonana/
└── README.md
```

## Tests

## Codes de Règles

### Règles de Base (Niveau 1)
- `C-L1` : Longueur de ligne (80 caractères max)
- `C-L2` : Lignes vides interdites
- `C-L3` : Indentation en TAB
- `C-L4` : Une variable par ligne
- `C-V1` : Déclarations en début de fonction
- `C-O1` : Nom de fichier snake_case
- `C-O2` : Maximum 3 fonctions par fichier
- `C-F1` : Nom de fonction snake_case
- `C-F2` : Nom de macro SCREAMING_SNAKE_CASE
- `C-F3` : Fonction 25 lignes max

### Règles Avancées (Niveau 2)
- `C-C1` : Format de commentaires
- `C-C2` : Commentaire de fonction obligatoire
- `C-G1` : Pas de globales non const
- `C-F4` : Maximum 4 paramètres
- `C-L5` : Pas de déclaration dans les boucles

## 🔧 Développement

### Tests
Le projet dispose d'une suite de tests complète avec **89.3%** de couverture :
```bash
# Lancer les tests
make test

# Avec couverture
go test -cover

# Avec rapport détaillé
go test -v -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### CI/CD
Une GitHub Action automatique exécute les tests à chaque push et pull request :
- Exécution de tous les tests
- Vérification de la couverture (minimum 85%)
- Compilation du projet
- Linter (golangci-lint)

Les pushs sont automatiquement rejetés si les tests échouent ou si la couverture descend sous 85%.

## License

Ce projet est sous licence MIT. Voir le fichier `LICENSE` pour plus de détails.

## Roadmap

- [x] Tests unitaires complets (89.3% coverage)
- [x] Intégration CI/CD (GitHub Actions)
- [ ] Option `--fix` pour corrections automatiques
- [ ] Support des fichiers de configuration
- [ ] Plugin VSCode
- [ ] Interface web
- [ ] Métriques de complexité
- [ ] Règles personnalisables

## Signaler un Bug

Si vous trouvez un bug, merci de créer une issue avec :
- Description du problème
- Fichier exemple qui cause le problème
- Version de Go utilisée
- Système d'exploitation

## Support

Pour toute question ou suggestion :
- Créer une issue sur GitHub
- Envoyer un email à : louis.malaval@epitech.eu

---

Développé pour la communauté Epitech