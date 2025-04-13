package main

import "testing"

func testIntegrationUserSchedule(t *testing.T) {
	staff1 := integrationTest.TestUsers[1]
	member1 := integrationTest.TestUsers[4]

	t.Run("APP_GROUP_SCHEDULE permission is required to create a personal schedule", func(t *testing.T) {
		_, err := postSchedule(member1.TestToken)
		if err == nil {
			t.Error("user was able to create schedule without APP_GROUP_SCHEDULE permissions")
		}
	})

	t.Run("user can create a personal schedule using a group schedule id", func(t *testing.T) {
		schedule, err := postSchedule(staff1.TestToken)
		if err != nil {
			t.Errorf("staff post schedule err %v", err)
		}

		integrationTest.UserSchedule = schedule
	})

	failCheck(t)
}
