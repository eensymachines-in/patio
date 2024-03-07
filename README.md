### Maiden tryst with Aquaponics:
---

400 litre fish tank, intentions to grow cherry tomatoes to understand the fine balanced nitrogen cycle, while keeping an eye open for possible avenues for Internet of things + automation.
A large part of successful aquaponics experimentation is all about diligent monitoring, taking corrective action on/before time & maintaining comprehensive logs of all the events and triggers with their consequential effects. This inspired me to build a small IoT project along as the experimental farm was being setup. 

- Water pump and its timing is vital to maitaining a the fine nitrogen cycle balance.
- Water pH has a small window in which the fish and plants coexists - too alkaline the plants wont respond, too acidic and the fish would die.
- Water temprature is instrumental in achieving fish health and apettite, while in a ceratin range of water temperature tomato plants are known to thrive.
- Water level loss is though minimal needs close monitoring for any sudden changes. 

Water pump timing isnt a issue if you have got the bell siphon working correctly. Bell siphons are riddled with problems of fine tuning. It would be not unless you have perfected this can you proceed for actual bacterial cultivation. Inlet mass flow of water (into the grow bed) has cascading effect on how the downstream flow rates get constrained.

 - Inlet flow is more than necessary to just start the siphon lock, would result in the air lock not getting a chance to open
 - Inlet flow is much less than necessaery would mean the air lock is not sustained and like a weir the water just negotiates what it can - siphon is never started.
  
What I found convenient is keeping the inlet mass flow a little more than necessary and as the siphon locks we can then time the motor to stop so that only under siphon lock and gravity water can flow out with a guarentee of lock opening at the desired level 

### Water pump control / timing :
----

1. Water pump raises the water from fish tank to the grow bed 
2. Grow bed head rises till bell siphon gets air locked 
3. An air locked siphon then starts draining water back to the fish tank 
4. Water head in the grow bed lowers will it reaches the slots 
5. Air lock is opened as air seeps into the siphon. - Drain stops
6. Water pump continues to raise the head in the grow bed till air lock is formed again

This cycle is what is referred to as the `Ebb-Flow` system