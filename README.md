# Purify the Cursed Frame

[![CI](https://github.com/itsabush1003/cursed-frame/actions/workflows/build-root.yml/badge.svg)](https://github.com/itsabush1003/cursed-frame/actions/workflows/build-root.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> **The mini game for small or medium team (about over 10 and under 60) to promote mutual understandings.**

!["タイトル画面"](images/Screenshot_Title.png)
!["ゲーム画面1"](images/Screenshot_Ingame.png)
!["ゲーム画面2"](images/Screenshot_Attack.png)

## Abstract

This game is a type of quiz game, but instead of general knowledge questions, the questions are about the other players in the game.

### Background

The idea for this game came about when I was a new employee. During a social gathering in my department, the department head asked us to come up with an entertainment event that would help foster camaraderie among the members, including us new employees.
At the time, I loved the quiz game that the company was releasing. The game featured a wizard protagonist, the player's avatar, who, along with several companions as cards, would share their power with them by answering quizzes, allowing them to attack enemies. This setting, along with the concept of "competition and cooperation" that the president had taught us during new employee training a few weeks prior to the department head's request, and the purpose of the meeting, all clicked into place in my mind.
Unfortunately, this project never materialized at the time. I simply lacked all the necessary resources, including time and my own technical skills.
Even so, I felt it would be a shame to abandon this project. The gears were simply clicking into place in my mind. So I decided to pursue this project as a personal project.

## Features

- The format of a browser game that is easy to participate in
  - Regardless of the type or specifications of your device, a certain degree of consistent operation can be guaranteed, and there's no need for the troublesome process of installing unofficial apps.
- Creating quizzes based on information provided by the users themselves
  - By actually conducting a survey with participants and using their responses as quiz options, we anticipate not only gaining personal information, but also fostering a sense of camaraderie from having answered the same survey, stimulating interest in options other than the correct answer, and increasing immersion through improved connectivity between the in-game world and the real world.
- Separation of the overall screen from individual screens
  - By utilizing the host-guest communication relationship, the overall information is displayed on the host's screen, and individual information is displayed on the guest's screen, eliminating the need to cram a large amount of information onto the guest's small screen.
- Competition and cooperation between teams
  - The overall goal was to create a system that emphasizes collaboration among all teams through a narrative, while also fostering a sense of competition by visualizing the contributions of individuals and each team.

## Tech Stacks

| Category | Technology                                    |
| :------- | :-------------------------------------------- |
| Frontend | React(v19), TypeScript, Emotion, Unity(WebGL) |
| Backend  | Go(1.25)                                      |
| Database | SQLite                                        |
| Library  | Connect, React-Unity-WebGL                    |
| CI/CD    | Github Actions                                |

### Why these technologies?

- React & TypeScript: It's the de facto standard configuration for web frontends and has a robust ecosystem, so it would be easier to implement interactive UIs.
- Emotion: For styling uses raw CSS, it allows for fine-tuning, and the design can be contained within each component.
- Unity(WebGL):　It's easy to create lightweight 3D games and output them as browser games.
- Go: The project was based on the premise of simultaneous communication by many users, so parallel processing was lightweight and easy to write, and in addition, cross-platform support was also easy.
- SQLite: Wanted to minimize the effort and external factors as much as possible, so choiced a storage solution that could be embedded in Golang.
- Connect: The connection between the frontend and backend can be unified through a common codebase, and server streaming, which is necessary for implementing core functionality, can be easily achieved.
- React-Unity-WebGL: Among the necessary screens, there were a certain number that weren't very game-like, such as admin panels, and using React allowed us to manage the UI with code.

## Architecture

```
[Browser] -> [React] <-[Connect/REST API]-> [Golang] <-> [SQLite]
                ^
                | [React-Unity-WebGL]
                v
              [Unity]
```

The detail of architecture is ~~[here](docs/architecture.md)~~(Under Construction).

## Usage

### Prerequisition

- Machine that can be used as a server publicly accessible on the internet at on-premises or cloud
  - ngrok
  - GCE
  - EC2
  - Google Cloud Run
  - AWS Lambda
  - Azure Container Apps
  - etc...
    - if you want to use deploy as container, image should be created by yourself

if you want to build from source (with customize)

- Unity ^2022.3
- bun ^1.3
- go ^1.25

### Setup

Download binary from releases

```bash
gh release download -R itsabush1003/cursed-frame
```

Or

```bash
curl -LO https://github.com/itsabush1003/cursed-frame/releases/latest/download/cursed_frame
```

Or Build from source on your machine

```bash
# clone repository
git clone https://github.com/itsabush1003/cursed-frame.git
cd cursed-frame

# build unity project
cd frontend/Unity/CursedFrame
your/unity/execution/path/Unity -quit -batchmode -nographic -projectpath . -buildTarget WebGL ../build/
# copy build artifacts to server usable space
cd ../build
cp WebGL/CursedFrame/Build/* ../../../backend/golang/dist/webgl/

cd ../../../

# build react project
cd frontend/react/cursed-frame
# change prefix of filename of webgl which built by Unity
echo "VITE_UNITY_WEBGL_NAME=CursedFrame" > .env.local

bun run build
# copy build artifacts to server usable space automatically
bun run deploy

cd ../../../

# build golang
cd backend/golang

# if you want to custom questions for participant,
# edit migration/Master/ProfileQuestion.csv before build

go build . -o cursed_frame

cd ../../
```

### Start Game

```bash
backend/golang/cursed_frame -N {total_participants_number} -T {separared_team_number} [-domain {domain_which_the_server_started}]
```

if you have started server, you can see 2 paths on the console by stdout, like

```
admin: <your_domain>:8888/[randam strings]/admin
guest: <your_domain>:8888/[random strings]/guest
```

First one which end with 'admin' is for you, as an administrator,
and last one which end with 'guest' is for participants,
so you should tell last one to participants, and must not tell first one.
At that time, don't forget to provide not only the path displayed on the screen, but also the domain name of the running server.
After that, you access the path, and click the button to start game, and then, just follow the on-screen instructions and you'll be fine.

if you want to know detail of screen transitions, you can look ~~[here](docs/screen_transitions.md)~~(Under Construction).

## Acknowledgement

- [React-Unity-WebGL](https://github.com/jeffreylanters/react-unity-webgl) - It's a fantastic library; without it, I wouldn't even have been able to start making this game.
- [Chibi_body](https://booth.pm/ja/items/5638034) - The 3D model of the player-assisted character, which we call the magician, is based on this model.
