package test

// sets of instructions to use in tests of other packages

// line can be Deposit:
// 	A (1): 10
// 	(deposit to A, TokenID 1, 10 units)
// or Transfer:
// 	A-B (1): 6 1
// 	(transfer from A to B, TokenID 1, 6 units, with fee 1)
// or Withdraw:
// 	A (1) E: 4
//  	exit to A, TokenID 1, 4 units)
// or NextBatch:
// 	> and here the comment
// 	move one batch forward

// Set0 has 3 batches, 29 different accounts, with:
// - 3 TokenIDs
// - 29+5+10 L1 txs (deposits & exits)
// - 21+53+7 L2 transactions
var SetTest0 = `
	// deposits TokenID: 1
	A (1): 50
	B (1): 5
	C (1): 20
	D (1): 25
	E (1): 25
	F (1): 25
	G (1): 25
	H (1): 25
	I (1): 25
	J (1): 25
	K (1): 25
	L (1): 25
	M (1): 25
	N (1): 25
	O (1): 25
	P (1): 25
	Q (1): 25
	R (1): 25
	S (1): 25
	T (1): 25
	U (1): 25
	V (1): 25
	W (1): 25
	X (1): 25
	Y (1): 25
	Z (1): 25

	// deposits TokenID: 2
	B (2): 5
	A (2): 20

	// deposits TokenID: 3
	B (3): 100

	// transactions TokenID: 1
	A-B (1): 5 1
	A-L (1): 10 1
	A-M (1): 5 1
	A-N (1): 5 1
	A-O (1): 5 1
	B-C (1): 3 1
	C-A (1): 3 255
	D-A (1): 5 1
	D-Z (1): 5 1
	D-Y (1): 5 1
	D-X (1): 5 1
	E-Z (1): 5 2
	E-Y (1): 5 1
	E-X (1): 5 1
	F-Z (1): 5 1
	G-K (1): 3 1
	G-K (1): 3 1
	G-K (1): 3 1
	H-K (1): 3 2
	H-K (1): 3 1
	H-K (1): 3 1

	> batch1

	// A (3) still does not exist, coordinator should create new L1Tx to create the account
	B-A (3): 5 1

	A-B (2): 5 1
	I-K (1): 3 1
	I-K (1): 3 1
	I-K (1): 3 1
	J-K (1): 3 1
	J-K (1): 3 1
	J-K (1): 3 1
	K-J (1): 3 1
	L-A (1): 5 1
	L-Z (1): 5 1
	L-Y (1): 5 1
	L-X (1): 5 1
	M-A (1): 5 1
	M-Z (1): 5 1
	M-Y (1): 5 1
	N-A (1): 5 1
	N-Z (1): 5 2
	N-Y (1): 5 1
	O-T (1): 3 1
	O-U (1): 3 1
	O-V (1): 3 1
	P-T (1): 3 1
	P-U (1): 3 1
	P-V (1): 3 5
	Q-O (1): 3 1
	Q-P (1): 3 1
	R-O (1): 3 1
	R-P (1): 3 1
	R-Q (1): 3 1
	S-O (1): 3 1
	S-P (1): 3 1
	S-Q (1): 3 1
	T-O (1): 3 1
	T-P (1): 3 1
	T-Q (1): 3 1
	U-Z (1): 5 3
	U-Y (1): 5 1
	U-T (1): 3 1
	V-Z (1): 5 0
	V-Y (1): 6 1
	V-T (1): 3 1
	W-K (1): 3 1
	W-J (1): 3 1
	W-A (1): 5 1
	W-Z (1): 5 1
	X-B (1): 5 1
	X-C (1): 5 50
	X-D (1): 5 1
	X-E (1): 5 1
	Y-B (1): 5 1
	Y-C (1): 5 1
	Y-D (1): 5 1
	Y-E (1): 5 1
	Z-A (1): 5 1

	// exits
	A (1) E: 5
	K (1) E: 5
	X (1) E: 5
	Y (1) E: 5
	Z (1) E: 5

	> batch2
	A (1): 50
	B (1): 5
	C (1): 20
	D (1): 25
	E (1): 25
	F (1): 25
	G (1): 25
	H (1): 25
	I (1): 25
	A-B (1): 5 1
	A-L (1): 10 1
	A-M (1): 5 1
	B-N (1): 5 1
	C-O (1): 5 1
	H-O (1): 5 1
	I-H (1): 5 1
	A (1) E: 5
`
