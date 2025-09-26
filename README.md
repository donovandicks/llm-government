# LLM Government

A multi-agent experiment to explore how LLMs may collaborate together to produce
structures and systems of government.

## TODO Items

- [X] Basic Agent-to-Agent communication
- [ ] Simulation Config
  - [X] Scenario
  - [ ] Population Settings
  - [ ] World Settings
- [ ] Agent Config
  - [ ] Agreeableness
  - [ ] Communication Tone & Style
  - [ ] Core Values & Beliefs
- [ ] Simulation Feedback
  - [ ] Tools for agents to actually apply policies
  - [ ] Subjects on which policies take effect
- [ ] Self Evolution
  - [ ] Explore if it's possible for agents to adjust the simulation itself
        either through config settings or modifying their own code

## Simulation

***WIP***

Ideally we want to have some simulation of "people" that the agents are governing.
The people need some indication of satisfaction with the government's policies,
and the policies need to have some effect on the people.

We can approach this in a video-gamey way by establishing a game loop where each
iteration every "player" takes some action or experiences some status effect. Players
will have some innate properties like age, health, hunger, and happiness, and also
items like money, food, etc. Actions may include buying goods, trading with other
players, etc. Policies may impact taxes, legality of certain goods or actions,
and other economic and social factors.
