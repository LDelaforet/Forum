# Forum

Projet FORUM développé en Go.

## Prérequis

- Go 1.23.0 ou supérieur
- MySQL
- Un navigateur web

## Installation

1. Clonez le repository :
```bash
git clone [URL_DU_REPO]
cd Forum
```

2. Configurez la base de données MySQL :
- Créez une base de données nommée `bdd_forum`
- Importez le fichier `bdd_forum.sql` dans phpMyAdmin :
  1. Ouvrez phpMyAdmin
  2. Sélectionnez la base de données `bdd_forum`
  3. Cliquez sur l'onglet "Importer"
  4. Sélectionnez le fichier `bdd_forum.sql`
  5. Cliquez sur "Exécuter"

3. Créez un fichier `.env` à la racine du projet avec les variables suivantes :
```env
# Configuration de la base de données
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_NAME=bdd_forum

# Configuration du serveur
HOST=localhost
PORT=8080

# Sécurité
SESSION_KEY=votre_clé_secrète_très_longue_et_complexe_ici
```

Pour générer une clé de session sécurisée, utilisez la commande suivante dans votre terminal :
```bash
openssl rand -base64 32
```
Copiez la sortie de cette commande comme valeur pour `SESSION_KEY` dans votre fichier `.env`.

4. Installez les dépendances :
```bash
cd src
go mod download
```

## Lancement du projet

1. Dans le dossier `src`, exécutez :
```bash
go run .
```

2. Ouvrez votre navigateur et accédez à `http://localhost:8080`

## Fonctionnalités

### Authentification
- Inscription de nouveaux utilisateurs
- Connexion/Déconnexion

### Posts
- Création de posts
- Modification de ses propres posts
- Suppression de ses propres posts
- Système de votes (plus/moins)
- Système de tags
- Recherche de posts

### Commentaires
- Ajout de commentaires aux posts
- Système de votes pour les commentaires
- Réponses aux commentaires (commentaires imbriqués)

### Interface
- Design moderne
- Thème sombre
- Navigation
- Barre de recherche

## Routes API

### Authentification
- `POST /register` - Inscription d'un nouvel utilisateur
- `POST /login` - Connexion d'un utilisateur
- `POST /logout` - Déconnexion

### Posts
- `GET /posts` - Liste des posts (paginée)
- `POST /posts` - Création d'un nouveau post
- `GET /posts/:id` - Détails d'un post
- `PUT /posts/:id` - Modification d'un post
- `DELETE /posts/:id` - Suppression d'un post
- `GET /posts/search` - Recherche de posts
- `POST /posts/:id/vote` - Vote sur un post

### Commentaires
- `GET /posts/:id/comments` - Liste des commentaires d'un post
- `POST /posts/:id/comments` - Ajout d'un commentaire
- `PUT /comments/:id` - Modification d'un commentaire
- `DELETE /comments/:id` - Suppression d'un commentaire
- `POST /comments/:id/vote` - Vote sur un commentaire

## Structure du projet

```
Forum/
├── src/
│   ├── config/         # Configuration et variables d'environnement
│   ├── controllers/    # Contrôleurs de l'application
│   ├── models/         # Modèles de données
│   ├── utils/          # Utilitaires et fonctions communes
│   ├── views/          # Templates HTML et assets statiques
│   │   ├── static/     # CSS, images, polices
│   │   └── *.html      # Templates HTML
│   ├── main.go         # Point d'entrée de l'application
│   └── go.mod          # Dépendances Go
├── .env               # Variables d'environnement
└── README.md
```

## Contribution

1. Fork le projet
2. Créez votre branche de fonctionnalité (`git checkout -b feature/AmazingFeature`)
3. Committez vos changements (`git commit -m 'Add some AmazingFeature'`)
4. Poussez vers la branche (`git push origin feature/AmazingFeature`)
5. Ouvrez une Pull Request

## Licence

Ce projet est sous licence MIT. Voir le fichier `LICENSE` pour plus de détails. 