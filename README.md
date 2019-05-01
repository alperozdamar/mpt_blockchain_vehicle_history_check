# Vehicle History Check
In Turkey, the second-hand car market is really huge. Approximately 6 million cars exchange between individuals yearly. Second hand car prices are directly related with car’s brand name, model year and feature of car. On the other hand, Turkish car buyers also care about some other features like car’s kilometer, car’s accident history. For instance, if there is an additional paint in the car,(other than the manufacturer’s paint in the factory) Turkish people don’t want to buy it in a regular second-hand price and car’s price reduces drastically. Car Mileage is also very important for Turkish people. They don’t want to pay much to used cars which has more than 200.000 km. As a result, some people (not few) manipulate car’s mileage. They lower car’s actual mileage to be able to sell their cars with a higher price. It is really hard for a regular buyer to understand this fraud. 
If every vehicle service in Turkey use a blockchain to keep service records of serviced car, Dishonest sellers can not fool buyers, easily. Blockchain is immutable, nobody can change car history and or data will never get lost. These records will be distributed to every car service in Turkey. (decentralization) If a buyer or anyone request a “history check” of a car, they can be able to see this immutable data. 

# Functionalities/ Success 

1-) Services add service information about incoming cars, successfully.

2-) Car service information distributed to all car services, successfully.

3-) Car service information can’t be changed. (Data Integrity is important.) 

4-) Car service information can be queried by users.

Please visit latest version of Project Plan: 

## What I have done so far? (DONE LIST)

|                |DONE							 |NEXT TODO|
|----------------|-------------------------------|-----------------------------|
|1-)		|Private/Public key generation, Sign/Verification |Integrate it with Heartbeats           
|2-)          |     Car Info. Insertion GUI      | Car Info. Query GUI         |
|3-)          |Define data structures for car info. , transaction |DONE!|

#### Transaction Object

|      Attribute          |Data Type|Description|
|----------------|-------------------------------|-----------------------------|
|transactionId   | String| UniqueId For Transaction           
|serviceId|String|Unique Id For Service (who creates the record)
|carPlate       |     String    | Unique Key of Car         |
|mileage      |int|	Car's mileage |
|insertDate|timestamp|	the date that car came to the service |
|transactionFee |int     |	fee for publishing the block |


####  Service GUI
Every Service has its own GUI. Lets Say Ankara has its own car service GUI, Istanbul has its own car service GUI.  

![For Car Gui Insert GUI:](https://github.com/alperozdamar/alperozdamar_mpt_project5/blob/master/car_gui_insert.png)


![For Car Gui Sequence Diagram:](https://github.com/alperozdamar/alperozdamar_mpt_project5/blob/master/Sequence_GUI.png)

####  RSA Public/Private Key Generation
I used crypto/rsa library to create public and private key pairs for every services who publishes car information to the blockchain. 

## What will I do next? (TODO LIST)

|                |NEXT TO DO
|-------------|------------------------------------------------------------|
|1-)		  |Signing hearbeat with Public Key of Receiver
|3-)          |     Verify received heartbeat with Private Key         
|3-)          |     Data distribution through peers(services)       
|4-)          |     Interface for users to read car information      
|5-)          |Test
|5-)          |Demo Video 
