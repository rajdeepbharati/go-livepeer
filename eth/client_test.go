package eth

import (
	"math/big"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	lpTypes "github.com/livepeer/go-livepeer/eth/types"
	"github.com/stretchr/testify/assert"
)

func copyTranscoders(transcoders []*lpTypes.Transcoder) []*lpTypes.Transcoder {
	cp := make([]*lpTypes.Transcoder, 0)
	for _, tr := range transcoders {
		trCp := new(lpTypes.Transcoder)
		trCp.Address = tr.Address
		trCp.DelegatedStake = new(big.Int)
		*trCp.DelegatedStake = *tr.DelegatedStake
		cp = append(cp, trCp)
	}
	return cp
}

func TestGetStakingHints(t *testing.T) {
	assert := assert.New(t)

	transcoders := []*lpTypes.Transcoder{
		{
			Address:        ethcommon.HexToAddress("aaa"),
			DelegatedStake: big.NewInt(5),
		},
		{
			Address:        ethcommon.HexToAddress("bbb"),
			DelegatedStake: big.NewInt(4),
		},
		{
			Address:        ethcommon.HexToAddress("ccc"),
			DelegatedStake: big.NewInt(3),
		},
		{
			Address:        ethcommon.HexToAddress("ddd"),
			DelegatedStake: big.NewInt(2),
		},
		{
			Address:        ethcommon.HexToAddress("eee"),
			DelegatedStake: big.NewInt(1),
		},
	}

	c := &client{}

	// oldDel = bbb
	// newDel = ccc
	// amount = 0
	// -> positions remain the same
	hints := c.getStakingHints(transcoders[2].Address, transcoders[1].Address, big.NewInt(0), transcoders)
	assert.Equal(hints.oldPosNext, transcoders[2].Address)
	assert.Equal(hints.oldPosPrev, transcoders[0].Address)
	assert.Equal(hints.newPosNext, transcoders[3].Address)
	assert.Equal(hints.newPosPrev, transcoders[1].Address)

	// oldDel = bbb
	// newDel = ccc
	// amount = 1
	// -> oldDel down 1 spot, newDel up one spot
	// copy the transcoders slice because getStakingHints actually mutates it
	tester0 := copyTranscoders(transcoders)
	hints = c.getStakingHints(transcoders[2].Address, transcoders[1].Address, big.NewInt(1), tester0)
	assert.Equal(hints.oldPosNext, transcoders[3].Address)
	assert.Equal(hints.oldPosPrev, transcoders[2].Address)
	assert.Equal(hints.newPosNext, transcoders[1].Address)
	assert.Equal(hints.newPosPrev, transcoders[0].Address)

	// same delegator (can be called as oldDel with zero value or olDel == newDel)
	// newdel = ccc
	// amount = 2
	// -> newDel up one spot
	// copy the transcoders slice because getStakingHints actually mutates it
	tester1 := copyTranscoders(transcoders)
	hints = c.getStakingHints(transcoders[2].Address, transcoders[2].Address, big.NewInt(2), tester1)
	assert.Equal(hints.oldPosNext, ethcommon.Address{})
	assert.Equal(hints.oldPosPrev, ethcommon.Address{})
	assert.Equal(hints.newPosNext, transcoders[1].Address)
	assert.Equal(hints.newPosPrev, transcoders[0].Address)

	// newDel = ddd
	// amount = 10
	// newDel becomes head
	// -> only newPosNext
	tester2 := copyTranscoders(transcoders)
	hints = c.getStakingHints(transcoders[3].Address, ethcommon.Address{}, big.NewInt(10), tester2)
	assert.Equal(hints.oldPosNext, ethcommon.Address{})
	assert.Equal(hints.oldPosPrev, ethcommon.Address{})
	assert.Equal(hints.newPosNext, transcoders[0].Address)
	assert.Equal(hints.newPosPrev, ethcommon.Address{})

	// newDel = eee
	// amount = 0
	// newdel is tail
	// -> only oldPosPrev
	tester3 := copyTranscoders(transcoders)
	hints = c.getStakingHints(transcoders[4].Address, ethcommon.Address{}, big.NewInt(0), tester3)
	assert.Equal(hints.oldPosNext, ethcommon.Address{})
	assert.Equal(hints.oldPosPrev, ethcommon.Address{})
	assert.Equal(hints.newPosNext, ethcommon.Address{})
	assert.Equal(hints.newPosPrev, transcoders[3].Address)

	// oldDel = eee
	// newDel = aaa
	// amount = 0
	// oldDel is tail
	// newDel is head
	// -> only newPosNext and oldPosPrev
	tester4 := copyTranscoders(transcoders)
	hints = c.getStakingHints(transcoders[0].Address, transcoders[4].Address, big.NewInt(0), tester4)
	assert.Equal(hints.oldPosNext, ethcommon.Address{})
	assert.Equal(hints.oldPosPrev, transcoders[3].Address)
	assert.Equal(hints.newPosNext, transcoders[1].Address)
	assert.Equal(hints.newPosPrev, ethcommon.Address{})

	// oldDel = aaa
	// newDel = eee
	// amount = 1
	// oldDel remains head
	// newDel remains tail
	// -> only newPosPrev and oldPosNext
	tester5 := copyTranscoders(transcoders)
	hints = c.getStakingHints(transcoders[4].Address, transcoders[0].Address, big.NewInt(1), tester5)
	assert.Equal(hints.oldPosNext, transcoders[1].Address)
	assert.Equal(hints.oldPosPrev, ethcommon.Address{})
	assert.Equal(hints.newPosNext, ethcommon.Address{})
	assert.Equal(hints.newPosPrev, transcoders[3].Address)
}
