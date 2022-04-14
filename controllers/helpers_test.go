package controllers

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	tmprbacv1 "rnemet.dev/temprolebindings/api/v1"
)

func Test_calculateCurrentStatus(t *testing.T) {
	testCases := []struct {
		name string
		trb  tmprbacv1.TempRoleBinding

		expected tmprbacv1.TempRoleBindingStatus
	}{
		{
			name: "Not set status",
			trb:  tmprbacv1.TempRoleBinding{},
			expected: tmprbacv1.TempRoleBindingStatus{
				Conditions: []tmprbacv1.Condition{
					{
						TransitionTime: metav1.Now(),
						Status:         true,
						Type:           tmprbacv1.TempRoleBindingStatusPending,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusDeclined,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusApproved,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusApplied,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusExpired,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusError,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusHold,
					},
				},
				Phase: tmprbacv1.TempRoleBindingStatusPending,
			},
		},
		{
			name: "Set status",
			trb: tmprbacv1.TempRoleBinding{
				Status: tmprbacv1.TempRoleBindingStatus{
					Conditions: []tmprbacv1.Condition{
						{
							TransitionTime: metav1.Now(),
							Status:         true,
							Type:           tmprbacv1.TempRoleBindingStatusPending,
						},
						{
							Status: false,
							Type:   tmprbacv1.TempRoleBindingStatusDeclined,
						},
						{
							TransitionTime: metav1.Now(),
							Status:         true,
							Type:           tmprbacv1.TempRoleBindingStatusApproved,
						},
						{
							Status: false,
							Type:   tmprbacv1.TempRoleBindingStatusApplied,
						},
						{
							Status: false,
							Type:   tmprbacv1.TempRoleBindingStatusExpired,
						},
						{
							Status: false,
							Type:   tmprbacv1.TempRoleBindingStatusError,
						},
						{
							Status: false,
							Type:   tmprbacv1.TempRoleBindingStatusHold,
						},
					},
					Phase: tmprbacv1.TempRoleBindingStatusPending,
				},
			},
			expected: tmprbacv1.TempRoleBindingStatus{
				Conditions: []tmprbacv1.Condition{
					{
						Status: true,
						Type:   tmprbacv1.TempRoleBindingStatusPending,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusDeclined,
					},
					{
						Status: true,
						Type:   tmprbacv1.TempRoleBindingStatusApproved,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusApplied,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusExpired,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusError,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusHold,
					},
				},
				Phase: tmprbacv1.TempRoleBindingStatusPending,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateCurrentStatus(tc.trb.Status.ToBaseStatus())

			for i, _ := range result.Conditions {
				assert.Equal(t, tc.expected.Conditions[i].Status, result.Conditions[i].Status)
				assert.Equal(t, tc.expected.Conditions[i].Type, result.Conditions[i].Type)
				if result.Conditions[i].Status {
					fmt.Printf("%v", result.Conditions[i])
					assert.True(t, result.Conditions[i].TransitionTime.After(metav1.Time{}.Time))
				}
			}

		})
	}
}

func Test_newStatus(t *testing.T) {
	testCases := []struct {
		name    string
		current tmprbacv1.BaseStatus
		next    tmprbacv1.RoleBindingStatus

		expected tmprbacv1.BaseStatus
	}{
		{
			name: "change pending 2 approved",
			current: tmprbacv1.BaseStatus{
				Conditions: []tmprbacv1.Condition{
					{
						TransitionTime: metav1.Now(),
						Status:         true,
						Type:           tmprbacv1.TempRoleBindingStatusPending,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusDeclined,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusApproved,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusApplied,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusExpired,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusError,
					},
				},
				Phase: tmprbacv1.TempRoleBindingStatusPending,
			},
			next: tmprbacv1.TempRoleBindingStatusApproved,

			expected: tmprbacv1.BaseStatus{
				Conditions: []tmprbacv1.Condition{
					{
						TransitionTime: metav1.Now(),
						Status:         true,
						Type:           tmprbacv1.TempRoleBindingStatusPending,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusDeclined,
					},
					{
						Status:         true,
						Type:           tmprbacv1.TempRoleBindingStatusApproved,
						TransitionTime: metav1.Now(),
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusApplied,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusExpired,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusError,
					},
				},
				Phase: tmprbacv1.TempRoleBindingStatusApproved,
			},
		},
		{
			name: "change approved 2 applied",
			current: tmprbacv1.BaseStatus{
				Conditions: []tmprbacv1.Condition{
					{
						TransitionTime: metav1.Now(),
						Status:         true,
						Type:           tmprbacv1.TempRoleBindingStatusPending,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusDeclined,
					},
					{
						TransitionTime: metav1.Now(),
						Status:         true,
						Type:           tmprbacv1.TempRoleBindingStatusApproved,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusApplied,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusExpired,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusError,
					},
				},
				Phase: tmprbacv1.TempRoleBindingStatusApproved,
			},
			next: tmprbacv1.TempRoleBindingStatusApplied,

			expected: tmprbacv1.BaseStatus{
				Conditions: []tmprbacv1.Condition{
					{
						TransitionTime: metav1.Now(),
						Status:         true,
						Type:           tmprbacv1.TempRoleBindingStatusPending,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusDeclined,
					},
					{
						TransitionTime: metav1.Now(),
						Status:         true,
						Type:           tmprbacv1.TempRoleBindingStatusApproved,
					},
					{
						TransitionTime: metav1.Now(),
						Status:         true,
						Type:           tmprbacv1.TempRoleBindingStatusApplied,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusExpired,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusError,
					},
				},
				Phase: tmprbacv1.TempRoleBindingStatusApplied,
			},
		},
		{
			name: "change applied 2 expired",
			current: tmprbacv1.BaseStatus{
				Conditions: []tmprbacv1.Condition{
					{
						TransitionTime: metav1.Now(),
						Status:         true,
						Type:           tmprbacv1.TempRoleBindingStatusPending,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusDeclined,
					},
					{
						TransitionTime: metav1.Now(),
						Status:         true,
						Type:           tmprbacv1.TempRoleBindingStatusApproved,
					},
					{
						TransitionTime: metav1.Now(),
						Status:         true,
						Type:           tmprbacv1.TempRoleBindingStatusApplied,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusExpired,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusError,
					},
				},
				Phase: tmprbacv1.TempRoleBindingStatusApplied,
			},
			next: tmprbacv1.TempRoleBindingStatusExpired,

			expected: tmprbacv1.BaseStatus{
				Conditions: []tmprbacv1.Condition{
					{
						TransitionTime: metav1.Now(),
						Status:         true,
						Type:           tmprbacv1.TempRoleBindingStatusPending,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusDeclined,
					},
					{
						TransitionTime: metav1.Now(),
						Status:         true,
						Type:           tmprbacv1.TempRoleBindingStatusApproved,
					},
					{
						TransitionTime: metav1.Now(),
						Status:         true,
						Type:           tmprbacv1.TempRoleBindingStatusApplied,
					},
					{
						TransitionTime: metav1.Now(),
						Status:         true,
						Type:           tmprbacv1.TempRoleBindingStatusExpired,
					},
					{
						Status: false,
						Type:   tmprbacv1.TempRoleBindingStatusError,
					},
				},
				Phase: tmprbacv1.TempRoleBindingStatusExpired,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resutl := newStatus(tc.current, tc.next)

			for i, _ := range resutl.Conditions {
				assert.Equal(t, resutl.Conditions[i].Status, tc.expected.Conditions[i].Status)
				assert.Equal(t, resutl.Conditions[i].Type, tc.expected.Conditions[i].Type)
				if resutl.Conditions[i].Status {
					assert.True(t, resutl.Conditions[i].TransitionTime.After(metav1.Time{}.Time))
				} else {
					assert.True(t, resutl.Conditions[i].TransitionTime.Equal(&metav1.Time{}))
				}
			}
		})
	}
}

func Test_holdOrApply(t *testing.T) {

	second, _ := time.ParseDuration("1s")
	testCases := []struct {
		name string
		trb  tmprbacv1.TempRoleBinding

		expected tmprbacv1.RoleBindingStatus
	}{
		{
			name: "duration on, and apply",
			trb: tmprbacv1.TempRoleBinding{
				Spec: tmprbacv1.TempRoleBindingSpec{
					Duration: metav1.Duration{
						Duration: second,
					},
				},
			},

			expected: tmprbacv1.TempRoleBindingStatusApplied,
		},
		{
			name: "no duration, and hold",
			trb: tmprbacv1.TempRoleBinding{
				Spec: tmprbacv1.TempRoleBindingSpec{
					StartStop: tmprbacv1.StartStop{
						From: metav1.Time{
							Time: time.Now().Add(time.Hour * 1),
						},
						To: metav1.Time{
							Time: time.Now().Add(time.Hour * 1),
						},
					},
				},
			},

			expected: tmprbacv1.TempRoleBindingStatusHold,
		},
		{
			name: "no duration, and apply",
			trb: tmprbacv1.TempRoleBinding{
				Spec: tmprbacv1.TempRoleBindingSpec{
					StartStop: tmprbacv1.StartStop{
						From: metav1.Time{
							Time: time.Now().Add(time.Hour * -1),
						},
						To: metav1.Time{
							Time: time.Now().Add(time.Hour * 1),
						},
					},
				},
			},

			expected: tmprbacv1.TempRoleBindingStatusApplied,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			next := holdOrApply(tmprbacv1.BaseSpec(tc.trb.Spec))

			assert.Equal(t, next, tc.expected)
		})
	}
}

func Test_isTempTempRoleBindingExpired(t *testing.T) {
	testCases := []struct {
		name string
		trb  tmprbacv1.TempRoleBinding

		expect bool
	}{
		{
			name: "Duration, not expired, not Applied",
			trb: tmprbacv1.TempRoleBinding{
				Spec: tmprbacv1.TempRoleBindingSpec{
					Duration: metav1.Duration{
						Duration: time.Nanosecond * 1,
					},
				},
			},
			expect: false,
		},
		{
			name: "Duration, expired, Applied",
			trb: tmprbacv1.TempRoleBinding{
				Spec: tmprbacv1.TempRoleBindingSpec{
					Duration: metav1.Duration{
						Duration: time.Nanosecond * 1,
					},
				},
				Status: tmprbacv1.TempRoleBindingStatus{
					Phase: tmprbacv1.TempRoleBindingStatusApplied,
					Conditions: []tmprbacv1.Condition{
						{
							TransitionTime: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
							Status: true,
							Type:   tmprbacv1.TempRoleBindingStatusApplied,
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "Duration, not expired, Applied",
			trb: tmprbacv1.TempRoleBinding{
				Spec: tmprbacv1.TempRoleBindingSpec{
					Duration: metav1.Duration{
						Duration: time.Hour * 1,
					},
				},
				Status: tmprbacv1.TempRoleBindingStatus{
					Phase: tmprbacv1.TempRoleBindingStatusApplied,
					Conditions: []tmprbacv1.Condition{
						{
							TransitionTime: metav1.Now(),
							Status:         true,
							Type:           tmprbacv1.TempRoleBindingStatusApplied,
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "StartStop, not expired, Applied",
			trb: tmprbacv1.TempRoleBinding{
				Spec: tmprbacv1.TempRoleBindingSpec{
					StartStop: tmprbacv1.StartStop{
						From: metav1.Now(),
						To: metav1.Time{
							Time: metav1.Now().Add(time.Hour * 1),
						},
					},
				},
				Status: tmprbacv1.TempRoleBindingStatus{
					Phase: tmprbacv1.TempRoleBindingStatusApplied,
					Conditions: []tmprbacv1.Condition{
						{
							TransitionTime: metav1.Now(),
							Status:         true,
							Type:           tmprbacv1.TempRoleBindingStatusApplied,
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "StartStop, expired, Applied",
			trb: tmprbacv1.TempRoleBinding{
				Spec: tmprbacv1.TempRoleBindingSpec{
					StartStop: tmprbacv1.StartStop{
						From: metav1.Now(),
						To: metav1.Time{
							Time: metav1.Now().Add(time.Microsecond * 1),
						},
					},
				},
				Status: tmprbacv1.TempRoleBindingStatus{
					Phase: tmprbacv1.TempRoleBindingStatusApplied,
					Conditions: []tmprbacv1.Condition{
						{
							TransitionTime: metav1.Now(),
							Status:         true,
							Type:           tmprbacv1.TempRoleBindingStatusApplied,
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "StartStop, not expired, not Applied",
			trb: tmprbacv1.TempRoleBinding{
				Spec: tmprbacv1.TempRoleBindingSpec{
					StartStop: tmprbacv1.StartStop{
						From: metav1.Now(),
						To: metav1.Time{
							Time: metav1.Now().Add(time.Microsecond * 1),
						},
					},
				},
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			time.Sleep(time.Millisecond * 10)
			assert.Equal(t, tc.expect, isTempTempRoleBindingExpired(tmprbacv1.BaseSpec(tc.trb.Spec), tc.trb.Status.ToBaseStatus()))
		})
	}
}

func Test_calculateNextStatus(t *testing.T) {
	testCases := []struct {
		name string
		trb  tmprbacv1.TempRoleBinding

		expected tmprbacv1.TempRoleBindingStatus
	}{
		{
			name: "Expired => Expired",
			trb: tmprbacv1.TempRoleBinding{
				Status: tmprbacv1.TempRoleBindingStatus{
					Conditions: []tmprbacv1.Condition{
						{
							TransitionTime: metav1.Now(),
							Status:         true,
							Type:           tmprbacv1.TempRoleBindingStatusExpired,
						},
					},
					Phase: tmprbacv1.TempRoleBindingStatusExpired,
				},
			},

			expected: tmprbacv1.TempRoleBindingStatus{
				Phase: tmprbacv1.TempRoleBindingStatusExpired,
			},
		},
		{
			name: "Declined => Declined",
			trb: tmprbacv1.TempRoleBinding{
				Status: tmprbacv1.TempRoleBindingStatus{
					Conditions: []tmprbacv1.Condition{
						{
							TransitionTime: metav1.Now(),
							Status:         true,
							Type:           tmprbacv1.TempRoleBindingStatusDeclined,
						},
					},
					Phase: tmprbacv1.TempRoleBindingStatusDeclined,
				},
			},

			expected: tmprbacv1.TempRoleBindingStatus{
				Phase: tmprbacv1.TempRoleBindingStatusDeclined,
			},
		},
		{
			name: "Applied => Applied (Duration)",
			trb: tmprbacv1.TempRoleBinding{
				Spec: tmprbacv1.TempRoleBindingSpec{
					Duration: metav1.Duration{
						Duration: time.Hour * 2,
					},
				},
				Status: tmprbacv1.TempRoleBindingStatus{
					Conditions: []tmprbacv1.Condition{
						{
							TransitionTime: metav1.Now(),
							Status:         true,
							Type:           tmprbacv1.TempRoleBindingStatusApplied,
						},
					},
					Phase: tmprbacv1.TempRoleBindingStatusApplied,
				},
			},

			expected: tmprbacv1.TempRoleBindingStatus{
				Phase: tmprbacv1.TempRoleBindingStatusApplied,
			},
		},
		{
			name: "Applied => Expired (Duration)",
			trb: tmprbacv1.TempRoleBinding{
				Spec: tmprbacv1.TempRoleBindingSpec{
					Duration: metav1.Duration{
						Duration: time.Nanosecond * 1,
					},
				},
				Status: tmprbacv1.TempRoleBindingStatus{
					Phase: tmprbacv1.TempRoleBindingStatusApplied,
					Conditions: []tmprbacv1.Condition{
						{
							TransitionTime: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
							Status: true,
							Type:   tmprbacv1.TempRoleBindingStatusApplied,
						},
						{
							Status: false,
							Type:   tmprbacv1.TempRoleBindingStatusExpired,
						},
					},
				},
			},

			expected: tmprbacv1.TempRoleBindingStatus{
				Phase: tmprbacv1.TempRoleBindingStatusExpired,
			},
		},
		{
			name: "Approved => Hold (StartStop)",
			trb: tmprbacv1.TempRoleBinding{
				Spec: tmprbacv1.TempRoleBindingSpec{
					StartStop: tmprbacv1.StartStop{
						From: metav1.Time{
							Time: time.Now().Add(2 * time.Hour),
						},
						To: metav1.Time{
							Time: time.Now().Add(5 * time.Hour),
						},
					},
				},
				Status: tmprbacv1.TempRoleBindingStatus{
					Phase: tmprbacv1.TempRoleBindingStatusApproved,
					Conditions: []tmprbacv1.Condition{
						{
							TransitionTime: metav1.Now(),
							Status:         true,
							Type:           tmprbacv1.TempRoleBindingStatusApproved,
						},
						{
							Status: false,
							Type:   tmprbacv1.TempRoleBindingStatusHold,
						},
					},
				},
			},

			expected: tmprbacv1.TempRoleBindingStatus{
				Phase: tmprbacv1.TempRoleBindingStatusHold,
			},
		},
		{
			name: "Approved => Applied (StartStop)",
			trb: tmprbacv1.TempRoleBinding{
				Spec: tmprbacv1.TempRoleBindingSpec{
					StartStop: tmprbacv1.StartStop{
						From: metav1.Time{
							Time: time.Now().Add(1 * time.Nanosecond),
						},
						To: metav1.Time{
							Time: time.Now().Add(5 * time.Hour),
						},
					},
				},
				Status: tmprbacv1.TempRoleBindingStatus{
					Phase: tmprbacv1.TempRoleBindingStatusApproved,
					Conditions: []tmprbacv1.Condition{
						{
							TransitionTime: metav1.Now(),
							Status:         true,
							Type:           tmprbacv1.TempRoleBindingStatusApproved,
						},
						{
							Status: false,
							Type:   tmprbacv1.TempRoleBindingStatusApplied,
						},
					},
				},
			},

			expected: tmprbacv1.TempRoleBindingStatus{
				Phase: tmprbacv1.TempRoleBindingStatusApplied,
			},
		},
		{
			name: "Approved => Applied (Duration)",
			trb: tmprbacv1.TempRoleBinding{
				Spec: tmprbacv1.TempRoleBindingSpec{
					Duration: metav1.Duration{
						Duration: time.Hour * 1,
					},
				},
				Status: tmprbacv1.TempRoleBindingStatus{
					Phase: tmprbacv1.TempRoleBindingStatusApproved,
					Conditions: []tmprbacv1.Condition{
						{
							TransitionTime: metav1.Now(),
							Status:         true,
							Type:           tmprbacv1.TempRoleBindingStatusApproved,
						},
						{
							TransitionTime: metav1.Now(),
							Status:         true,
							Type:           tmprbacv1.TempRoleBindingStatusPending,
						},
						{
							Status: false,
							Type:   tmprbacv1.TempRoleBindingStatusApplied,
						},
					},
				},
			},

			expected: tmprbacv1.TempRoleBindingStatus{
				Phase: tmprbacv1.TempRoleBindingStatusApplied,
			},
		},
		{
			name: "Hold => Hold (StartStop)",
			trb: tmprbacv1.TempRoleBinding{
				Spec: tmprbacv1.TempRoleBindingSpec{
					StartStop: tmprbacv1.StartStop{
						From: metav1.Time{
							Time: time.Now().Add(1 * time.Hour),
						},
						To: metav1.Time{
							Time: time.Now().Add(5 * time.Hour),
						},
					},
				},
				Status: tmprbacv1.TempRoleBindingStatus{
					Phase: tmprbacv1.TempRoleBindingStatusHold,
					Conditions: []tmprbacv1.Condition{
						{
							TransitionTime: metav1.Now(),
							Status:         true,
							Type:           tmprbacv1.TempRoleBindingStatusHold,
						},
						{
							Status: false,
							Type:   tmprbacv1.TempRoleBindingStatusApplied,
						},
					},
				},
			},

			expected: tmprbacv1.TempRoleBindingStatus{
				Phase: tmprbacv1.TempRoleBindingStatusHold,
			},
		},
		{
			name: "Hold => Applied (StartStop)",
			trb: tmprbacv1.TempRoleBinding{
				Spec: tmprbacv1.TempRoleBindingSpec{
					StartStop: tmprbacv1.StartStop{
						From: metav1.Time{
							Time: time.Now().Add(1 * time.Nanosecond),
						},
						To: metav1.Time{
							Time: time.Now().Add(5 * time.Hour),
						},
					},
				},
				Status: tmprbacv1.TempRoleBindingStatus{
					Phase: tmprbacv1.TempRoleBindingStatusHold,
					Conditions: []tmprbacv1.Condition{
						{
							TransitionTime: metav1.Now(),
							Status:         true,
							Type:           tmprbacv1.TempRoleBindingStatusHold,
						},
						{
							Status: false,
							Type:   tmprbacv1.TempRoleBindingStatusApplied,
						},
					},
				},
			},

			expected: tmprbacv1.TempRoleBindingStatus{
				Phase: tmprbacv1.TempRoleBindingStatusApplied,
			},
		},
		{
			name: "Pending => Automatic, Approved",
			trb: tmprbacv1.TempRoleBinding{
				Status: tmprbacv1.TempRoleBindingStatus{
					Phase: tmprbacv1.TempRoleBindingStatusPending,
					Conditions: []tmprbacv1.Condition{
						{
							TransitionTime: metav1.Now(),
							Status:         true,
							Type:           tmprbacv1.TempRoleBindingStatusPending,
						},
						{
							Status: false,
							Type:   tmprbacv1.TempRoleBindingStatusApproved,
						},
					},
				},
			},

			expected: tmprbacv1.TempRoleBindingStatus{
				Phase: tmprbacv1.TempRoleBindingStatusApproved,
			},
		},
		{
			name: "Pending => Manual Approved",
			trb: tmprbacv1.TempRoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						tmprbacv1.StatusAnnotation: "Approved",
					},
				},
				Spec: tmprbacv1.TempRoleBindingSpec{
					ApprovalRequired: true,
				},
				Status: tmprbacv1.TempRoleBindingStatus{
					Phase: tmprbacv1.TempRoleBindingStatusPending,
					Conditions: []tmprbacv1.Condition{
						{
							TransitionTime: metav1.Now(),
							Status:         true,
							Type:           tmprbacv1.TempRoleBindingStatusPending,
						},
						{
							Status: false,
							Type:   tmprbacv1.TempRoleBindingStatusApproved,
						},
					},
				},
			},

			expected: tmprbacv1.TempRoleBindingStatus{
				Phase: tmprbacv1.TempRoleBindingStatusApproved,
			},
		},
		{
			name: "Pending => Manual Declined",
			trb: tmprbacv1.TempRoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						tmprbacv1.StatusAnnotation: "Declined",
					},
				},
				Spec: tmprbacv1.TempRoleBindingSpec{
					ApprovalRequired: true,
				},
				Status: tmprbacv1.TempRoleBindingStatus{
					Phase: tmprbacv1.TempRoleBindingStatusPending,
					Conditions: []tmprbacv1.Condition{
						{
							TransitionTime: metav1.Now(),
							Status:         true,
							Type:           tmprbacv1.TempRoleBindingStatusPending,
						},
						{
							Status: false,
							Type:   tmprbacv1.TempRoleBindingStatusDeclined,
						},
					},
				},
			},

			expected: tmprbacv1.TempRoleBindingStatus{
				Phase: tmprbacv1.TempRoleBindingStatusDeclined,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			time.Sleep(time.Millisecond * 10)
			_, next := calculateNextStatus(tc.trb)
			assert.Equal(t, tc.expected.Phase, next.Phase)
		})
	}
}
