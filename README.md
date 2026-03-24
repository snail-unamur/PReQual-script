# PReQual - Pull Request Dataset Analyzer

## Installation

### 1 - Sonar CLI Runner

```bash
# Pulling the docker image
docker pull  sonarsource/sonar-scanner-cli
```

The docker container will be launch automatically by the script during the analysis.

### 2 - Setup a Docker network

```bash
docker network create prequal-net
```

### 3 - SonarQube

```bash
# Pulling the docker image
docker pull sonarqube

# Run the image
docker run -d --name prequal-sonarqube -e SONAR_ES_BOOTSTRAP_CHECKS_DISABLE=true -p 9000:9000 --network=prequal-sonar-net sonarqube:latest
```

Once the image is running, the admin password must be reset by going on Sonar's [main page](http://localhost:9000). 
The web interface run on `http://localhost:9000`, but it will be different if you run the analysis on a distant server.

### 4 - Setup Token access

Once SonarQube is running, a global access token must be set:

- Got to `My Account` then to `security`
- Create a new token by selecting `Global Analysis Token`. 
- Copy the newly created token, and paste it in the .env file of the analyzer root folder (a `.env.example` is available for reference).

### 5 - Setup MongoDB
```bash
# Pulling the docker image
docker pull mongo:latest

# Run the image
docker run -d --name mongo -p 27017:27017 --network=prequal-net mongo:latest
```

### 6 - Setup .env file

Use the `.env.example` to setup the parameter to run the script

## Run the script locally

To run the analyzer script:

```bash
go run main.go -repos organization/repository 
```

Two arguments can be passed: 
- `-repos organization/repository(,organization/repository)*`, to precise on which GitHub repositorys the analysis will be run.
  - `organization` is the GitHub organization name
  - `repository` is the GitHub repository name
- [Option] `-workspace my-workspace`, to precise the workspace folder for the analysis, the default workspace is `./tmp` in the root script folder.
- [Option] `-metrics flag1,flag2,...`, to precise which SonarQube flags should be included in the analysis, the default is `Cognitive complexity` and `Cyclomatic complexity`. The available flags are listed in the next section, the extended list of flags can be found on SonarQube's official documentation.


### Sonar Flags

The SonarQube analysis can be customized by passing flags to the script.
- Cyclomatic complexity -> `complexity`
- Cognitive complexity -> `cognitive_complexity`
- Code smells -> `code_smells`
- Development cost -> `development_cost`
- Duplicated lines -> `duplicated_lines`
- Number of lines of code -> `ncloc`
- Software quality maintainability rating -> `software_quality_maintainability_rating`
