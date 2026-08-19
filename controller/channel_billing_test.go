package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPickDeepSeekBalanceIndex(t *testing.T) {
	tests := []struct {
		name         string
		balanceInfos []DeepSeekBalanceInfo
		want         int
	}{
		{
			name: "国内版账号返回 CNY 条目",
			balanceInfos: []DeepSeekBalanceInfo{
				{Currency: "CNY", TotalBalance: "100.00"},
			},
			want: 0,
		},
		{
			name: "国际版账号只返回 USD 条目时回退第一条",
			balanceInfos: []DeepSeekBalanceInfo{
				{Currency: "USD", TotalBalance: "50.00"},
			},
			want: 0,
		},
		{
			name: "多币种条目时优先取 CNY",
			balanceInfos: []DeepSeekBalanceInfo{
				{Currency: "USD", TotalBalance: "50.00"},
				{Currency: "CNY", TotalBalance: "100.00"},
			},
			want: 1,
		},
		{
			name:         "无任何条目时返回 -1",
			balanceInfos: []DeepSeekBalanceInfo{},
			want:         -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pickDeepSeekBalanceIndex(tt.balanceInfos)
			assert.Equal(t, tt.want, got)
		})
	}
}
