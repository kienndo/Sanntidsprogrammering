# TTK4145 - ELEVATOR PROJECT, GROUP 63

## Functionality
There are some factors to take into consideration with this code. We have implemented a system with master-slave topology, and implemented process pairs for primary and backup in case the master shuts down. Therefore, the first computer that runs the program will become the master and run the primary process, and the other computers that has to be run X seconds after the first computer, will become the backups. By killing the master terminal, currently will both other computers take over as primary. Ideally, only one of them would become the primary and the other a backup, but the logic is still there. 


## Notes regarding the code
We were unsuccessful in implementing the code for taking care of disconnections. We therefore implement the pseudo code here to maintain the code quality in the original code.


# TO DO LIST
- [X] 

