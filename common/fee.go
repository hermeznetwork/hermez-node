package common

import (
	"fmt"
	"math"
	"math/big"

	"github.com/hermeznetwork/tracerr"
)

// MaxFeePlan is the maximum value of the FeePlan
const MaxFeePlan = 256

// FeePlan represents the fee model, a position in the array indicates the
// percentage of tokens paid in concept of fee for a transaction
var FeePlan = [MaxFeePlan]float64{}

// FeeFactorLsh60 is the feeFactor << 60
var FeeFactorLsh60 [256]*big.Int

// RecommendedFee is the recommended fee to pay in USD per transaction set by
// the coordinator according to the tx type (if the tx requires to create an
// account and register, only register or he account already esists)
type RecommendedFee struct {
	ExistingAccount        float64 `json:"existingAccount"`
	CreatesAccount         float64 `json:"createAccount"`
	CreatesAccountInternal float64 `json:"createAccountInternal"`
}

// FeeSelector is used to select a percentage from the FeePlan.
type FeeSelector uint8

// Percentage returns the associated percentage of the FeeSelector
func (f FeeSelector) Percentage() float64 {
	if f == 0 {
		return 0
	} else if f < 32 { //nolint:gomnd
		return math.Pow(2, -60.0+float64(f)*(-8.0-(-60.0))/32.0) //nolint:gomnd
	} else if f < 192 { //nolint:gomnd
		return math.Pow(2, -8.0+(float64(f)-32.0)*(0.0-(-8.0))/160.0) //nolint:gomnd
	} else {
		return math.Pow(2, (float64(f) - 192.0)) //nolint:gomnd
	}
}

// CalcFeeAmount calculates the fee amount in tokens from an amount and
// feeSelector (fee index).
func CalcFeeAmount(amount *big.Int, feeSel FeeSelector) (*big.Int, error) {
	feeAmount := new(big.Int).Mul(amount, FeeFactorLsh60[int(feeSel)])
	if feeSel < 192 { //nolint:gomnd
		feeAmount.Rsh(feeAmount, 60)
	}
	if feeAmount.BitLen() > 128 { //nolint:gomnd
		return nil, tracerr.Wrap(fmt.Errorf("FeeAmount overflow (feeAmount doesn't fit in 128 bits)"))
	}
	return feeAmount, nil
}

func init() {
	setFeeFactorLsh60(&FeeFactorLsh60)
}

//nolint:gomnd
func setFeeFactorLsh60(feeFactorLsh60 *[256]*big.Int) {
	feeFactorLsh60[0], _ = new(big.Int).SetString("0", 10)
	feeFactorLsh60[1], _ = new(big.Int).SetString("3", 10)
	feeFactorLsh60[2], _ = new(big.Int).SetString("9", 10)
	feeFactorLsh60[3], _ = new(big.Int).SetString("29", 10)
	feeFactorLsh60[4], _ = new(big.Int).SetString("90", 10)
	feeFactorLsh60[5], _ = new(big.Int).SetString("279", 10)
	feeFactorLsh60[6], _ = new(big.Int).SetString("861", 10)
	feeFactorLsh60[7], _ = new(big.Int).SetString("2655", 10)
	feeFactorLsh60[8], _ = new(big.Int).SetString("8192", 10)
	feeFactorLsh60[9], _ = new(big.Int).SetString("25267", 10)
	feeFactorLsh60[10], _ = new(big.Int).SetString("77935", 10)
	feeFactorLsh60[11], _ = new(big.Int).SetString("240387", 10)
	feeFactorLsh60[12], _ = new(big.Int).SetString("741455", 10)
	feeFactorLsh60[13], _ = new(big.Int).SetString("2286960", 10)
	feeFactorLsh60[14], _ = new(big.Int).SetString("7053950", 10)
	feeFactorLsh60[15], _ = new(big.Int).SetString("21757357", 10)
	feeFactorLsh60[16], _ = new(big.Int).SetString("67108864", 10)
	feeFactorLsh60[17], _ = new(big.Int).SetString("206992033", 10)
	feeFactorLsh60[18], _ = new(big.Int).SetString("638450708", 10)
	feeFactorLsh60[19], _ = new(big.Int).SetString("1969251187", 10)
	feeFactorLsh60[20], _ = new(big.Int).SetString("6074000999", 10)
	feeFactorLsh60[21], _ = new(big.Int).SetString("18734780191", 10)
	feeFactorLsh60[22], _ = new(big.Int).SetString("57785961645", 10)
	feeFactorLsh60[23], _ = new(big.Int).SetString("178236271212", 10)
	feeFactorLsh60[24], _ = new(big.Int).SetString("549755813888", 10)
	feeFactorLsh60[25], _ = new(big.Int).SetString("1695678735018", 10)
	feeFactorLsh60[26], _ = new(big.Int).SetString("5230188203117", 10)
	feeFactorLsh60[27], _ = new(big.Int).SetString("16132105731538", 10)
	feeFactorLsh60[28], _ = new(big.Int).SetString("49758216191607", 10)
	feeFactorLsh60[29], _ = new(big.Int).SetString("153475319327371", 10)
	feeFactorLsh60[30], _ = new(big.Int).SetString("473382597799226", 10)
	feeFactorLsh60[31], _ = new(big.Int).SetString("1460111533771401", 10)
	feeFactorLsh60[32], _ = new(big.Int).SetString("4503599627370496", 10)
	feeFactorLsh60[33], _ = new(big.Int).SetString("4662418725241772", 10)
	feeFactorLsh60[34], _ = new(big.Int).SetString("4826838566504035", 10)
	feeFactorLsh60[35], _ = new(big.Int).SetString("4997056660946426", 10)
	feeFactorLsh60[36], _ = new(big.Int).SetString("5173277483525749", 10)
	feeFactorLsh60[37], _ = new(big.Int).SetString("5355712719992597", 10)
	feeFactorLsh60[38], _ = new(big.Int).SetString("5544581521179432", 10)
	feeFactorLsh60[39], _ = new(big.Int).SetString("5740110766256133", 10)
	feeFactorLsh60[40], _ = new(big.Int).SetString("5942535335269230", 10)
	feeFactorLsh60[41], _ = new(big.Int).SetString("6152098391292193", 10)
	feeFactorLsh60[42], _ = new(big.Int).SetString("6369051672525772", 10)
	feeFactorLsh60[43], _ = new(big.Int).SetString("6593655794699191", 10)
	feeFactorLsh60[44], _ = new(big.Int).SetString("6826180564135515", 10)
	feeFactorLsh60[45], _ = new(big.Int).SetString("7066905301857248", 10)
	feeFactorLsh60[46], _ = new(big.Int).SetString("7316119179121470", 10)
	feeFactorLsh60[47], _ = new(big.Int).SetString("7574121564787630", 10)
	feeFactorLsh60[48], _ = new(big.Int).SetString("7841222384935199", 10)
	feeFactorLsh60[49], _ = new(big.Int).SetString("8117742495163242", 10)
	feeFactorLsh60[50], _ = new(big.Int).SetString("8404014066019092", 10)
	feeFactorLsh60[51], _ = new(big.Int).SetString("8700380982019120", 10)
	feeFactorLsh60[52], _ = new(big.Int).SetString("9007199254740992", 10)
	feeFactorLsh60[53], _ = new(big.Int).SetString("9324837450483544", 10)
	feeFactorLsh60[54], _ = new(big.Int).SetString("9653677133008070", 10)
	feeFactorLsh60[55], _ = new(big.Int).SetString("9994113321892852", 10)
	feeFactorLsh60[56], _ = new(big.Int).SetString("10346554967051498", 10)
	feeFactorLsh60[57], _ = new(big.Int).SetString("10711425439985194", 10)
	feeFactorLsh60[58], _ = new(big.Int).SetString("11089163042358864", 10)
	feeFactorLsh60[59], _ = new(big.Int).SetString("11480221532512266", 10)
	feeFactorLsh60[60], _ = new(big.Int).SetString("11885070670538460", 10)
	feeFactorLsh60[61], _ = new(big.Int).SetString("12304196782584386", 10)
	feeFactorLsh60[62], _ = new(big.Int).SetString("12738103345051544", 10)
	feeFactorLsh60[63], _ = new(big.Int).SetString("13187311589398382", 10)
	feeFactorLsh60[64], _ = new(big.Int).SetString("13652361128271030", 10)
	feeFactorLsh60[65], _ = new(big.Int).SetString("14133810603714496", 10)
	feeFactorLsh60[66], _ = new(big.Int).SetString("14632238358242940", 10)
	feeFactorLsh60[67], _ = new(big.Int).SetString("15148243129575260", 10)
	feeFactorLsh60[68], _ = new(big.Int).SetString("15682444769870398", 10)
	feeFactorLsh60[69], _ = new(big.Int).SetString("16235484990326484", 10)
	feeFactorLsh60[70], _ = new(big.Int).SetString("16808028132038184", 10)
	feeFactorLsh60[71], _ = new(big.Int).SetString("17400761964038240", 10)
	feeFactorLsh60[72], _ = new(big.Int).SetString("18014398509481984", 10)
	feeFactorLsh60[73], _ = new(big.Int).SetString("18649674900967100", 10)
	feeFactorLsh60[74], _ = new(big.Int).SetString("19307354266016140", 10)
	feeFactorLsh60[75], _ = new(big.Int).SetString("19988226643785704", 10)
	feeFactorLsh60[76], _ = new(big.Int).SetString("20693109934102996", 10)
	feeFactorLsh60[77], _ = new(big.Int).SetString("21422850879970388", 10)
	feeFactorLsh60[78], _ = new(big.Int).SetString("22178326084717744", 10)
	feeFactorLsh60[79], _ = new(big.Int).SetString("22960443065024532", 10)
	feeFactorLsh60[80], _ = new(big.Int).SetString("23770141341076920", 10)
	feeFactorLsh60[81], _ = new(big.Int).SetString("24608393565168772", 10)
	feeFactorLsh60[82], _ = new(big.Int).SetString("25476206690103088", 10)
	feeFactorLsh60[83], _ = new(big.Int).SetString("26374623178796784", 10)
	feeFactorLsh60[84], _ = new(big.Int).SetString("27304722256542060", 10)
	feeFactorLsh60[85], _ = new(big.Int).SetString("28267621207428992", 10)
	feeFactorLsh60[86], _ = new(big.Int).SetString("29264476716485880", 10)
	feeFactorLsh60[87], _ = new(big.Int).SetString("30296486259150520", 10)
	feeFactorLsh60[88], _ = new(big.Int).SetString("31364889539740816", 10)
	feeFactorLsh60[89], _ = new(big.Int).SetString("32470969980652968", 10)
	feeFactorLsh60[90], _ = new(big.Int).SetString("33616056264076368", 10)
	feeFactorLsh60[91], _ = new(big.Int).SetString("34801523928076480", 10)
	feeFactorLsh60[92], _ = new(big.Int).SetString("36028797018963968", 10)
	feeFactorLsh60[93], _ = new(big.Int).SetString("37299349801934200", 10)
	feeFactorLsh60[94], _ = new(big.Int).SetString("38614708532032280", 10)
	feeFactorLsh60[95], _ = new(big.Int).SetString("39976453287571408", 10)
	feeFactorLsh60[96], _ = new(big.Int).SetString("41386219868205992", 10)
	feeFactorLsh60[97], _ = new(big.Int).SetString("42845701759940776", 10)
	feeFactorLsh60[98], _ = new(big.Int).SetString("44356652169435488", 10)
	feeFactorLsh60[99], _ = new(big.Int).SetString("45920886130049064", 10)
	feeFactorLsh60[100], _ = new(big.Int).SetString("47540282682153840", 10)
	feeFactorLsh60[101], _ = new(big.Int).SetString("49216787130337544", 10)
	feeFactorLsh60[102], _ = new(big.Int).SetString("50952413380206176", 10)
	feeFactorLsh60[103], _ = new(big.Int).SetString("52749246357593568", 10)
	feeFactorLsh60[104], _ = new(big.Int).SetString("54609444513084120", 10)
	feeFactorLsh60[105], _ = new(big.Int).SetString("56535242414857984", 10)
	feeFactorLsh60[106], _ = new(big.Int).SetString("58528953432971760", 10)
	feeFactorLsh60[107], _ = new(big.Int).SetString("60592972518301040", 10)
	feeFactorLsh60[108], _ = new(big.Int).SetString("62729779079481632", 10)
	feeFactorLsh60[109], _ = new(big.Int).SetString("64941939961305936", 10)
	feeFactorLsh60[110], _ = new(big.Int).SetString("67232112528152736", 10)
	feeFactorLsh60[111], _ = new(big.Int).SetString("69603047856152960", 10)
	feeFactorLsh60[112], _ = new(big.Int).SetString("72057594037927936", 10)
	feeFactorLsh60[113], _ = new(big.Int).SetString("74598699603868352", 10)
	feeFactorLsh60[114], _ = new(big.Int).SetString("77229417064064608", 10)
	feeFactorLsh60[115], _ = new(big.Int).SetString("79952906575142816", 10)
	feeFactorLsh60[116], _ = new(big.Int).SetString("82772439736411984", 10)
	feeFactorLsh60[117], _ = new(big.Int).SetString("85691403519881552", 10)
	feeFactorLsh60[118], _ = new(big.Int).SetString("88713304338870912", 10)
	feeFactorLsh60[119], _ = new(big.Int).SetString("91841772260098192", 10)
	feeFactorLsh60[120], _ = new(big.Int).SetString("95080565364307680", 10)
	feeFactorLsh60[121], _ = new(big.Int).SetString("98433574260675088", 10)
	feeFactorLsh60[122], _ = new(big.Int).SetString("101904826760412352", 10)
	feeFactorLsh60[123], _ = new(big.Int).SetString("105498492715187056", 10)
	feeFactorLsh60[124], _ = new(big.Int).SetString("109218889026168304", 10)
	feeFactorLsh60[125], _ = new(big.Int).SetString("113070484829715968", 10)
	feeFactorLsh60[126], _ = new(big.Int).SetString("117057906865943520", 10)
	feeFactorLsh60[127], _ = new(big.Int).SetString("121185945036602080", 10)
	feeFactorLsh60[128], _ = new(big.Int).SetString("125459558158963264", 10)
	feeFactorLsh60[129], _ = new(big.Int).SetString("129883879922611968", 10)
	feeFactorLsh60[130], _ = new(big.Int).SetString("134464225056305472", 10)
	feeFactorLsh60[131], _ = new(big.Int).SetString("139206095712305920", 10)
	feeFactorLsh60[132], _ = new(big.Int).SetString("144115188075855872", 10)
	feeFactorLsh60[133], _ = new(big.Int).SetString("149197399207736800", 10)
	feeFactorLsh60[134], _ = new(big.Int).SetString("154458834128129216", 10)
	feeFactorLsh60[135], _ = new(big.Int).SetString("159905813150285632", 10)
	feeFactorLsh60[136], _ = new(big.Int).SetString("165544879472823968", 10)
	feeFactorLsh60[137], _ = new(big.Int).SetString("171382807039763104", 10)
	feeFactorLsh60[138], _ = new(big.Int).SetString("177426608677741952", 10)
	feeFactorLsh60[139], _ = new(big.Int).SetString("183683544520196384", 10)
	feeFactorLsh60[140], _ = new(big.Int).SetString("190161130728615360", 10)
	feeFactorLsh60[141], _ = new(big.Int).SetString("196867148521350176", 10)
	feeFactorLsh60[142], _ = new(big.Int).SetString("203809653520824704", 10)
	feeFactorLsh60[143], _ = new(big.Int).SetString("210996985430374272", 10)
	feeFactorLsh60[144], _ = new(big.Int).SetString("218437778052336608", 10)
	feeFactorLsh60[145], _ = new(big.Int).SetString("226140969659431936", 10)
	feeFactorLsh60[146], _ = new(big.Int).SetString("234115813731887040", 10)
	feeFactorLsh60[147], _ = new(big.Int).SetString("242371890073204160", 10)
	feeFactorLsh60[148], _ = new(big.Int).SetString("250919116317926528", 10)
	feeFactorLsh60[149], _ = new(big.Int).SetString("259767759845223936", 10)
	feeFactorLsh60[150], _ = new(big.Int).SetString("268928450112610944", 10)
	feeFactorLsh60[151], _ = new(big.Int).SetString("278412191424611840", 10)
	feeFactorLsh60[152], _ = new(big.Int).SetString("288230376151711744", 10)
	feeFactorLsh60[153], _ = new(big.Int).SetString("298394798415473600", 10)
	feeFactorLsh60[154], _ = new(big.Int).SetString("308917668256258432", 10)
	feeFactorLsh60[155], _ = new(big.Int).SetString("319811626300571264", 10)
	feeFactorLsh60[156], _ = new(big.Int).SetString("331089758945647936", 10)
	feeFactorLsh60[157], _ = new(big.Int).SetString("342765614079526208", 10)
	feeFactorLsh60[158], _ = new(big.Int).SetString("354853217355483904", 10)
	feeFactorLsh60[159], _ = new(big.Int).SetString("367367089040392768", 10)
	feeFactorLsh60[160], _ = new(big.Int).SetString("380322261457230720", 10)
	feeFactorLsh60[161], _ = new(big.Int).SetString("393734297042700352", 10)
	feeFactorLsh60[162], _ = new(big.Int).SetString("407619307041649408", 10)
	feeFactorLsh60[163], _ = new(big.Int).SetString("421993970860748544", 10)
	feeFactorLsh60[164], _ = new(big.Int).SetString("436875556104673216", 10)
	feeFactorLsh60[165], _ = new(big.Int).SetString("452281939318863872", 10)
	feeFactorLsh60[166], _ = new(big.Int).SetString("468231627463774080", 10)
	feeFactorLsh60[167], _ = new(big.Int).SetString("484743780146408320", 10)
	feeFactorLsh60[168], _ = new(big.Int).SetString("501838232635853056", 10)
	feeFactorLsh60[169], _ = new(big.Int).SetString("519535519690447872", 10)
	feeFactorLsh60[170], _ = new(big.Int).SetString("537856900225221888", 10)
	feeFactorLsh60[171], _ = new(big.Int).SetString("556824382849223680", 10)
	feeFactorLsh60[172], _ = new(big.Int).SetString("576460752303423488", 10)
	feeFactorLsh60[173], _ = new(big.Int).SetString("596789596830947200", 10)
	feeFactorLsh60[174], _ = new(big.Int).SetString("617835336512516864", 10)
	feeFactorLsh60[175], _ = new(big.Int).SetString("639623252601142528", 10)
	feeFactorLsh60[176], _ = new(big.Int).SetString("662179517891295872", 10)
	feeFactorLsh60[177], _ = new(big.Int).SetString("685531228159052416", 10)
	feeFactorLsh60[178], _ = new(big.Int).SetString("709706434710967808", 10)
	feeFactorLsh60[179], _ = new(big.Int).SetString("734734178080785536", 10)
	feeFactorLsh60[180], _ = new(big.Int).SetString("760644522914461440", 10)
	feeFactorLsh60[181], _ = new(big.Int).SetString("787468594085400704", 10)
	feeFactorLsh60[182], _ = new(big.Int).SetString("815238614083298816", 10)
	feeFactorLsh60[183], _ = new(big.Int).SetString("843987941721497088", 10)
	feeFactorLsh60[184], _ = new(big.Int).SetString("873751112209346432", 10)
	feeFactorLsh60[185], _ = new(big.Int).SetString("904563878637727744", 10)
	feeFactorLsh60[186], _ = new(big.Int).SetString("936463254927548160", 10)
	feeFactorLsh60[187], _ = new(big.Int).SetString("969487560292816640", 10)
	feeFactorLsh60[188], _ = new(big.Int).SetString("1003676465271706112", 10)
	feeFactorLsh60[189], _ = new(big.Int).SetString("1039071039380895744", 10)
	feeFactorLsh60[190], _ = new(big.Int).SetString("1075713800450443776", 10)
	feeFactorLsh60[191], _ = new(big.Int).SetString("1113648765698447360", 10)
	feeFactorLsh60[192], _ = new(big.Int).SetString("1", 10)
	feeFactorLsh60[193], _ = new(big.Int).SetString("2", 10)
	feeFactorLsh60[194], _ = new(big.Int).SetString("4", 10)
	feeFactorLsh60[195], _ = new(big.Int).SetString("8", 10)
	feeFactorLsh60[196], _ = new(big.Int).SetString("16", 10)
	feeFactorLsh60[197], _ = new(big.Int).SetString("32", 10)
	feeFactorLsh60[198], _ = new(big.Int).SetString("64", 10)
	feeFactorLsh60[199], _ = new(big.Int).SetString("128", 10)
	feeFactorLsh60[200], _ = new(big.Int).SetString("256", 10)
	feeFactorLsh60[201], _ = new(big.Int).SetString("512", 10)
	feeFactorLsh60[202], _ = new(big.Int).SetString("1024", 10)
	feeFactorLsh60[203], _ = new(big.Int).SetString("2048", 10)
	feeFactorLsh60[204], _ = new(big.Int).SetString("4096", 10)
	feeFactorLsh60[205], _ = new(big.Int).SetString("8192", 10)
	feeFactorLsh60[206], _ = new(big.Int).SetString("16384", 10)
	feeFactorLsh60[207], _ = new(big.Int).SetString("32768", 10)
	feeFactorLsh60[208], _ = new(big.Int).SetString("65536", 10)
	feeFactorLsh60[209], _ = new(big.Int).SetString("131072", 10)
	feeFactorLsh60[210], _ = new(big.Int).SetString("262144", 10)
	feeFactorLsh60[211], _ = new(big.Int).SetString("524288", 10)
	feeFactorLsh60[212], _ = new(big.Int).SetString("1048576", 10)
	feeFactorLsh60[213], _ = new(big.Int).SetString("2097152", 10)
	feeFactorLsh60[214], _ = new(big.Int).SetString("4194304", 10)
	feeFactorLsh60[215], _ = new(big.Int).SetString("8388608", 10)
	feeFactorLsh60[216], _ = new(big.Int).SetString("16777216", 10)
	feeFactorLsh60[217], _ = new(big.Int).SetString("33554432", 10)
	feeFactorLsh60[218], _ = new(big.Int).SetString("67108864", 10)
	feeFactorLsh60[219], _ = new(big.Int).SetString("134217728", 10)
	feeFactorLsh60[220], _ = new(big.Int).SetString("268435456", 10)
	feeFactorLsh60[221], _ = new(big.Int).SetString("536870912", 10)
	feeFactorLsh60[222], _ = new(big.Int).SetString("1073741824", 10)
	feeFactorLsh60[223], _ = new(big.Int).SetString("2147483648", 10)
	feeFactorLsh60[224], _ = new(big.Int).SetString("4294967296", 10)
	feeFactorLsh60[225], _ = new(big.Int).SetString("8589934592", 10)
	feeFactorLsh60[226], _ = new(big.Int).SetString("17179869184", 10)
	feeFactorLsh60[227], _ = new(big.Int).SetString("34359738368", 10)
	feeFactorLsh60[228], _ = new(big.Int).SetString("68719476736", 10)
	feeFactorLsh60[229], _ = new(big.Int).SetString("137438953472", 10)
	feeFactorLsh60[230], _ = new(big.Int).SetString("274877906944", 10)
	feeFactorLsh60[231], _ = new(big.Int).SetString("549755813888", 10)
	feeFactorLsh60[232], _ = new(big.Int).SetString("1099511627776", 10)
	feeFactorLsh60[233], _ = new(big.Int).SetString("2199023255552", 10)
	feeFactorLsh60[234], _ = new(big.Int).SetString("4398046511104", 10)
	feeFactorLsh60[235], _ = new(big.Int).SetString("8796093022208", 10)
	feeFactorLsh60[236], _ = new(big.Int).SetString("17592186044416", 10)
	feeFactorLsh60[237], _ = new(big.Int).SetString("35184372088832", 10)
	feeFactorLsh60[238], _ = new(big.Int).SetString("70368744177664", 10)
	feeFactorLsh60[239], _ = new(big.Int).SetString("140737488355328", 10)
	feeFactorLsh60[240], _ = new(big.Int).SetString("281474976710656", 10)
	feeFactorLsh60[241], _ = new(big.Int).SetString("562949953421312", 10)
	feeFactorLsh60[242], _ = new(big.Int).SetString("1125899906842624", 10)
	feeFactorLsh60[243], _ = new(big.Int).SetString("2251799813685248", 10)
	feeFactorLsh60[244], _ = new(big.Int).SetString("4503599627370496", 10)
	feeFactorLsh60[245], _ = new(big.Int).SetString("9007199254740992", 10)
	feeFactorLsh60[246], _ = new(big.Int).SetString("18014398509481984", 10)
	feeFactorLsh60[247], _ = new(big.Int).SetString("36028797018963968", 10)
	feeFactorLsh60[248], _ = new(big.Int).SetString("72057594037927936", 10)
	feeFactorLsh60[249], _ = new(big.Int).SetString("144115188075855872", 10)
	feeFactorLsh60[250], _ = new(big.Int).SetString("288230376151711744", 10)
	feeFactorLsh60[251], _ = new(big.Int).SetString("576460752303423488", 10)
	feeFactorLsh60[252], _ = new(big.Int).SetString("1152921504606846976", 10)
	feeFactorLsh60[253], _ = new(big.Int).SetString("2305843009213693952", 10)
	feeFactorLsh60[254], _ = new(big.Int).SetString("4611686018427387904", 10)
	feeFactorLsh60[255], _ = new(big.Int).SetString("9223372036854775808", 10)
}
