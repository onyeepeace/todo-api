# JotIt: A Collaborative Productivity Application

## Overview
JotIt is a complete productivity application that allows users to create, manage, and share multiple notes. The system supports collaborative editing, fine-grained access control, and real-time synchronization across devices. I worked on this project whilst on the CBF x Monzo mentorship program. Coming into the program, I had a goal of diving into distributed systems, systems design and backend architecture. I wanted to make a switch from Frontend to Backend Engineering.

## Features
- Designed and built a high-performance backend in Go and PostgreSQL, optimised for low-latency and high-concurrency, processing thousands of requests per second.
- Implemented role-based access control (RBAC) and fine-grained authorization to manage collaborative editing and secure data access.
- Implemented concurrency control mechanisms to ensure consistency across distributed clients for collaborative editing and explored CRDTs for live collaborative editing, leveraging WebSockets.
- Set up monitoring and observability using TimescaleDB, Grafana, and structured logging, enabling real-time performance tracking and incident response.
- Deployed on AWS (EC2, RDS, ALB) with automated CI/CD pipelines, ensuring seamless releases with zero downtime


<br>

> [!TIP]
> I gave a presentation about this at the close-out session of the CBF X Monzo mentorship program, which you can find [here](https://docs.google.com/presentation/d/e/2PACX-1vRgKpa_8NP_Kw0NyXLvAHw3IGXw4XMxdj0-ug--WP5V3Q8cUcaN7aO_c1eWNIoXGAbenHDQeeva4VFd/pub?start=false&loop=false&delayms=3000). This details my process from start to finish with fun and interesting videos and emojis to describe my experience. Some things to look out for:
> - The use of emojis on the data model slides. This was the most challenging part of the project, and I used emojis to convey how I felt at each point where I had to rethink and redesign the project based on new ideas or refactoring.
> - The use of videos to convey my feelings when I implemented some features.
> - If you do enjoy the presentation or would like to give me some feedback or chat about this project or what I'm up to, do leave me a message on [LinkedIn](https://www.linkedin.com/in/onyeepeace/)

<br>

# Below is a more detailed implementation process of this project
<br>

## Strategic Goals
1. Provide a robust, secure platform for personal and team productivity
2. Enable seamless collaboration through real-time updates
3. Ensure data integrity through proper concurrency control
4. Scale efficiently to support growing user base and data volume

### Product Requirement document
This was a key part of this project, not just as a way to showcase product thinking and being product-focused, but it was vital in designing the features of the app and thinking about extensibility. Having this document helped in making decisions and understanding tradeoffs.
<img width="439" alt="Screenshot 2025-03-31 at 11 42 30" src="https://github.com/user-attachments/assets/b88fdef0-ff06-46ec-aa10-7fc144069962" /> <img width="439" alt="Screenshot 2025-03-31 at 11 42 30" src="https://github.com/user-attachments/assets/b88fdef0-ff06-46ec-aa10-7fc144069962" />

## Data Model
This started as a simple todo list and progressed into a data model to support multiple lists, sharing, and permissions. Progressing through each design, brought about a mindset shift in the way I thought about users and scale. At this point was when I realised how much abstraction from the backend the frontend provides. I went from having a global todo list to a todo list for a specific user.
<img width="568" alt="Screenshot 2025-03-13 at 11 16 46" src="https://github.com/user-attachments/assets/6f114ec4-ca5c-4fbb-ad98-01e559a3e5d9" />

The next iteration was having a folder of todos. Basically lists of todos (I did change the name from Lists to Items as it became confusing to have lists and todo lists).
<img width="568" alt="Screenshot 2025-03-13 at 11 52 55" src="https://github.com/user-attachments/assets/e4db00cb-8304-4e45-8200-2ffc90cf545a" />

I started to think about other possibilities. What if a user doesn't want todos but plain text or a different data type?

<img width="465" alt="Screenshot 2025-03-13 at 12 42 15" src="https://github.com/user-attachments/assets/e142febe-43f4-40cd-af8c-fef138eaca4d" /> <img width="465" alt="Screenshot 2025-03-13 at 12 42 15" src="https://github.com/user-attachments/assets/72b2f840-fb77-471a-b6cf-1d11ce23ee83" />

I came up with the option to have an item as either a note or a todo.

<img width="465" alt="Screenshot 2025-03-13 at 12 42 15" src="https://github.com/user-attachments/assets/5cbd3ec7-4f17-4185-af6c-5331225e565d" />

After some more consideration and reading, the final data model allowed users to have either a note or todo list in an item.
<img width="732" alt="Screenshot 2025-03-13 at 12 52 03" src="https://github.com/user-attachments/assets/7bd3592c-075d-4927-8a27-a9bb7d2bd43e" />


## API Design
Designing the API endpoints was a moment where I had to think about the app and understand tradeoffs of the different ways I could have done it. Having the product requirement document to refer back to was also crucial here. For example, when it came to thinking about the resource IDs, I could do it two ways:
- A scoped ID: This meant, search could be performed on todos inside parent folders and blocked off the possibility to search globally. It could still be done but would involve complex queries.
- A global ID: This allowed searching todos across the whole app, regardless of folders.
With the search feature in mind, I could make informed decision even though it was a nice-to-have feature.

<img width="460" alt="Screenshot 2025-03-31 at 21 21 00" src="https://github.com/user-attachments/assets/8aa8904b-64a3-4e72-b081-aa0db92213d8" /> <img width="475" alt="Screenshot 2025-03-31 at 21 22 41" src="https://github.com/user-attachments/assets/901d90b8-2422-41ea-be70-84f77fd76b00" />

The API follows RESTful principles with clear resource hierarchies:

### Lists Endpoints
```
GET /api/lists → Fetches all lists
POST /api/lists → Creates a new list
PUT /api/lists/{listId} → Edits the list name
DELETE /api/lists/{listId} → Deletes list and associated todos
```

### Todos Endpoints
```
GET /api/lists/{listId}/todos → Get all todos in a list
POST /api/lists/{listId}/todos → Create a todo in a list
PUT /api/lists/{listId}/todos/{todoId} → Edit a todo
PATCH /api/lists/{listId}/todos/{todoId} → Mark a todo as done
DELETE /api/lists/{listId}/todos/{todoId} → Delete a todo
```

> [!Note]
> I took the AWS Certified Cloud Practitioner exam and passed 🥳. Studying for the exam and passing it was pivotal for this project. I was able to set up AWS services (EC2, RDS, ALB, Route53) to get the project together. The learning was immediately useful and it was exciting having to set up these services (ofcourse, I struggled but I figured it out).

## Authentication
- OAuth2 for authentication with identity providers
- JWT for stateless authentication

## Authorization
Role-based access control (RBAC) for fine-grained permissions

<img width="458" alt="Screenshot 2025-03-13 at 15 25 50" src="https://github.com/user-attachments/assets/18370cf0-db0c-49bf-a2ab-3c839d5fb909" /> <img width="530" alt="Screenshot 2025-03-13 at 12 52 31" src="https://github.com/user-attachments/assets/19fbe493-352f-4df0-bd97-9e1e64d224d9" />

<img width="545" alt="Screenshot 2025-04-01 at 09 31 14" src="https://github.com/user-attachments/assets/3010bee4-75a7-4f41-baca-2ec16f75dad4" />


## Concurrency Control
To handle concurrent updates, I implemented optimistic concurrency control. Users could edit the same item at the same time but not the same row. Editing the same row at the same time led to a 409 conflict and I used reject on concurrent update to handle this.

<img width="699" alt="Screenshot 2025-03-14 at 15 35 30" src="https://github.com/user-attachments/assets/32f46f1c-91be-4cf4-8658-76ed1ccaf28a" />

## System Architecture
Jotit follows a monolith architecture designed for scalability and maintainability:

### Key Components
- **Frontend**: React-based SPA with real-time updates
- **API Layer**: Go-based RESTful API with proper error handling and validation
- **Authentication**: OAuth2 with JWT for stateless authentication
- **Database**: PostgreSQL with optimized schemas for performance
- **Metrics**: TimescaleDB and Grafana stack for monitoring and alerting

## Security Measures
- HTTPS for all connections
- JWT for Authn
- RBAC for Authz
- CORS configuration for frontend security
<br>

# Productionisation
## Monitoring and Observability
This was the most exciting part of this project. Seeing the graphs in Grafana was just beautiful.
- Request metrics collection via TimescaleDB: I used this as it was an easy integration on top of PostgreSQL the primary database. It was also free to run and I didn't want to incur expenses with other paid tools since I wanted to play around and understand how it worked. (I exhausted the free tier from AWS and I was billed. Going forward, I opted for free tools or free trials)
- Custom application metrics for business-critical paths
- Grafana dashboards for visualization
- Structured logging with correlation IDs
- Error tracking and alerting

### Current setup
<img width="617" alt="Screenshot 2025-04-01 at 10 54 50" src="https://github.com/user-attachments/assets/b434144e-c77c-4553-931d-0a99ef1f3924" />

### Other options considered
<img width="703" alt="Screenshot 2025-04-01 at 10 55 00" src="https://github.com/user-attachments/assets/b0e83f12-7278-45f8-a6a9-2f718578821a" />

### Deployment
<img width="594" alt="Screenshot 2025-04-01 at 10 55 15" src="https://github.com/user-attachments/assets/f436d155-197f-406b-843c-70021d13d07d" />

## Development Workflow
- Gitflow branching strategy
- CI/CD pipeline with GitHub Actions
- Docker for consistent development environments

## Deployment
- AWS infrastructure managed via Docker

## Future Roadmap
- Live collaboration using CRDT and WebSockets
- Smart lists with filtering and sorting
- Integration with third-party services (calendar, email)

## Lessons learnt
- The mindset shift from thinking in the perspective of the frontend vs the backend is huge. On the frontend, you mostly work in the context of a ready-made user. On the backend, it's always multi-user and the concept of a user has to be defined (via Authn, Authz, roles). You can't assume a ready-made user. 
- You get a lot of value from thinking about things and designing it, and when it comes to building, you may want to do something different.
- I very quickly realised, the code doesn't matter if the design is not great. You'd end up with lots of rewrites and wasted time.
- I learned that a well-thought-out design is more important than jumping straight into writing code. At first, I started coding right away, but I quickly realized that constant design changes made my initial effort inefficient. I had to rethink my approach. On the frontend, proof of work is often the code you write or the UI you show. The system design is already done, and APIs are provided. But on the backend, I had to shift my mindset: the foundation matters more than the code itself. Once I embraced this, I spent more time thinking, reading, and understanding what I wanted to build.
As a result, I wrote less code, but more meaningful code. Now, with a clear design in place, I’m integrating tests confidently, knowing that new features can be built and validated without major rewrites.
- Evolution over revolution.
- Reinforced my desire and passion for backend engineering.
- Reading engineering blogs to understand how complex systems are built. I really like how detailed the Monzo ones are.
- Reading books and articles.

## Reading and Writing
As with learning new things, I spent a lot of time reading. A couple of the books I read (recommended by my mentor):
  - Designing Data Intensive Applications
  - Seven databases in seven weeks
  - The pragmatic programmer
  - How to read a book
  - How to take smart notes (This is a must read. You only need chapter 1 to start implementing. Absolute game changer.)
I write on my [blog](https://substack.com/@peaceonyehanere?utm_source=user-menu) and on [LinkedIn](https://www.linkedin.com/in/onyeepeace/) on new things I learn or implement.

## What next?
1. I'm working on implementing load testing. I read about performance testing, and even though there aren't a lot of users to use this app (it turns out I can simulate it), understanding its importance is key for me.
    ### Why:
      Setting up monitoring was the most exciting part of this project. Seeing the graphs up on Grafana wasn't just fascinating but intriguing, and it led to a couple things happening:
    - My initial excitement died down a bit as I wanted to see more traction from more users (users I didn't have 😂).
    - I wanted to see the graphs on Grafana with more traction.
    - It led me to start thinking about query optimisation, failure/success rate, and code optimisation.
    ### How:
      I have done some reading and exploration, and here's what I'm doing:
      - Utilising K6 from Grafana (I already tried it out, and it works) 
      - I read an article about using Goroutines to understand the principles behind the tools built for load testing (I'm excited to try this out)
        
2. As I learn new things, I would update this document with my implementation.
3. With hindsight from this project, I'm working on designing a different system and would share that too.

## Conclusion
If you got to this point, you would have noticed I'm really excited about understanding how system design works and the tools that are used in this domain. I also read up on new things, listen to talks and go down the deep end of things I find interesting. This was a lot to do in 4 months coming from understanding basic CRUD functionality. Of course, without leaving out the guidance from my amazing mentor. There were times when it became overwhelming learning new things and we would spend time unravelling my thoughts and understanding why I was thinking a certain way. Not forgetting all the resources he shared that were extremely useful. Looking back at my start point, I can laugh about the way I thought of things or implemented them 😂. I have grown a lot and it is so rewarding and interesting.
