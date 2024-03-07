### Maiden tryst with Aquaponics:
---

400 litre fish tank, intentions to grow cherry tomatoes to understand the fine balanced nitrogen cycle, while keeping an eye open for possible avenues for Internet of things + automation.
A large part of successful aquaponics experimentation is all about diligent monitoring, taking corrective action on/before time & maintaining comprehensive logs of all the events and triggers with their consequential effects. This inspired me to build a small IoT project along as the experimental farm was being setup. 

- Water pump and its timing is vital to maitaining a the fine nitrogen cycle balance.
- Water pH has a narrow window in which the __fish and plants coexists__ : too alkaline the plants won't respond, too acidic and the fish would die.
- Water temprature is instrumental in achieving fish health and apettite, while only in a ceratin range of water temperature tomato plants are known to thrive.
- Water level loss is though minimal needs close monitoring for any sudden changes. 

Water pump __timing__ isn't considered if you have got the bell siphon working correctly. Bell siphons are riddled with problems of fine tuning. It would be not __unless you have perfected this can you proceed for actual bacterial cultivation.__ 

Inlet mass flow of water (into the grow bed) has cascading effect on how the downstream flow rates get constrained.

 - _Inlet flow is more than necessary_ to just start the siphon lock, would result in the air lock not getting a chance to open. Thus growbed water reaches an equilibrium at the slots and water just free flows downstream to the fish. Unless of course you time switch the pump off and break the lock.
 - _Inlet flow is much less than necessery_ would mean the air lock is not sustained and like a weir the water just negotiates extra head just above the opening of the siphon. - __siphon is never started.__ For lower inflows the siphon needs a re-design and hence an expensive proposition.
  
What I found convenient (and economical) is keeping the inlet mass flow a little more than necessary and as the siphon locks we can then time the motor to stop so that only under siphon lock and gravity water can flow out with a guarentee of lock opening at the desired level 

### Water cycle - Flood & Drain :
----

1. Water pump raises the water from fish tank to the grow bed 
2. Growbed head rises till bell siphon gets air locked 
3. An air locked siphon then starts draining water back to the fish tank (against gravity) 
4. Water head in the grow bed lowers will it reaches the slots (on the bell of siphon 2.5" from the bottom)
5. Air lock is opened as air seeps into the siphon. - Drain flow stops
6. Water pump continues to raise the head in the grow bed till air lock is formed again

This cycle is what is referred to as the `Ebb-Flow` system. An automated way to flood and drain the grow bed is achieved with the siphon or siphon+ pump timing. Barring the bottom 2.5" of the grow bed which is always flooded, the fertile area remains damp and thus aids growth of bacteria. 
Such bacteria help sustain the nitrogen cycle. 

### Inflow just above required ! Pump is operation timed:
-----

#### Lets consider the pump ON 24x7:

When the pump is on for complete ebb-flow cycle the inflow does have an influence on 

1. Time it takes to drain 
2. Breaking the siphon lock

While being drained if the inflow > outlfow, the growbed would overflow, while if the inflow is lesser < outflow the growbed would be steadily drain out. Time required to drain the same amount of water though in such a case would be more than if the inflow was zero. So in the ideal world the inflow should be a square wave. 

1. High possible inflow till the grow bed is flooded 
2. Siphon is locked the inflow should be minimal as possible

Lower inflows tend to create difficulties in intiating / sustaining the siphon while higher inflows tend to make stronger siphon locks which are difficult to break. This setup needs the siphon to initiate & break smoothly. Hence we have a narrow window of inflow rate to control. Valves though are simplest / cheapest to use in such case the iteration to reach the exact inflow is quite tedious plus valves over period of time do accumulate grime that changes the inflow unexpectedly.

#### Pump has timing / sensor control:

While iterating the inflow, one can time the flooding and drain cycles. A clock driven relay can then sweetly time the relay such that exactly when the siphon starts the pump stops while when the siphon lock is opened the pump motor can resume.

Another interesting way out is attaching a [water flow sensor](https://www.amazon.in/Water-Flow-Sensor-by-Robokart/dp/B00ZNAXNRO/ref=sr_1_5?sr=8-5) which can detect outflow. Outflow indicates siphon lock initiation while when the flow stops it indicates an opened siphon lock and thus can signal relay to resume operation.