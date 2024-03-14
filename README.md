# TTK4145 - ELEVATOR PROJECT
## GROUP 63:
- Ingrid Kristine Bøe @ingriKBo
- Kien Ninh Do @kiennd
- Siri Helene Wahl @sirihelenew

## Usage
### Installation
### Dependencies
### Build and Run

## Project Description
**Create software for controlling `n` elevators working in parallel across `m` floors**, with the given main requirements
1. **No calls are lost**
2. **The lights and buttons should function as expected, with  appropriate feedback to users when they press them** (the button lights are a service guarantee).
3. **The door should function as expected**, with a doorlight triggered at reasonable times and a obstruction sensor functioning properly. 
4. **An individual elevator should behave sensibly and efficiently**. It should only stop at floors with relevant requests and adjust its direction based on these. 

and the given secondary requirement

   1. **Calls should be served as efficiently as possible**.

## Solution
Our code is written in `Golang`, which was a decision based on its efficiency for concurrent programming. The main features of our code are presented below:

### Functionality
In our solution, we have implemented a system with a **master-slave topology**. We use **process pairs** for primary and backup roles in case the master shuts down. 

The first computer that runs the program will become the master and run the primary process. Meanwhile, the other computers become backups, starting *X* seconds after the first computer. 

Currently, if the master terminal is killed, both other computers take over as primary. Ideally, only one of them should become the primary and the other a backup, but we were unsuccessful in implementing this.

### Notes regarding the code
We were unable to successfully implement the code for handling disconnections. For this reason, we have included pseudo code to maintain the code quality in the original code.
