package elevatorGovernor

import (
	"fmt"

	. "github.com/perkjelsvik/TTK4145-sanntid/project/constants"
	esm "github.com/perkjelsvik/TTK4145-sanntid/project/elevatorStateMachine"
	hw "github.com/perkjelsvik/TTK4145-sanntid/project/hardware"
)

//NOTE: queue suggestion so far
/*{0,1,0} {0,0,0} {0,1,0}
{0,0,0} {0,1,0} {0,1,0}
{0,0,1} {0,0,1} {1,0,0}
{1,0,0} {1,0,0} {0,1,0}*/

func GOV_init(ID int, ch esm.Channels) {
	fmt.Println("until we print")
	hw.SetStopLamp(1)
	var queue = [NumFloors][NumButtons][NumElevators]int{}
	fmt.Println(queue)
	var press Keypress

	for floor := 0; floor < NumFloors; floor++ {
		for btn := 0; btn < NumButtons; btn++ {
			if queue[floor][ID][btn] == 1 {
				press.Floor = floor
				press.Btn = Button(btn)
				ch.NewOrderChan <- press
			}
		}
	}

}

/*NOTE: should we have a compare between new network queue
and already existing queue? only forward new orders to esm
if actually new order? Could be done by for example:
newQueue[floor][btn] != queue[floor][btn]
*/
