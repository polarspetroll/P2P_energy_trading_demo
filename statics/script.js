
var units = {
	"hour": 1,
	"day": 24,
	"week": 24 * 7,
	"month": 24 * 30
}

function calculate() {
	var period   = Number(document.getElementById('dur').value)
	var unit     = units[document.getElementById('timeunits').value]
	var quantity = Number(document.getElementById('quantity').value)
	var out      = document.getElementById('out')
	final_out    = period * unit * quantity // example formule
	out.removeAttribute('hidden')
	out.innerText = `total price : ${final_out}$`

}