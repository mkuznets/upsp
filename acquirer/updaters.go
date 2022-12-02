package acquirer

import (
	"log"
	"time"
)

func (a *acquirerImpl) asyncRefunder() {
	for {
		payments, err := a.s.List(PaymentStateConfirmed)
		if err != nil {
			log.Printf("[ERR] could not list payments: %v", err)
		}
		log.Println(payments)

		for _, payment := range payments {
			if shouldRefund(payment.CardNumber) {
				_, err = a.CancelPayment(payment.Id, payment.Version)
				if err != nil {
					log.Printf("[ERR] failed to refund payment %s: %s", payment.Id, err)
				}
			}
		}

		time.Sleep(10 * time.Second)
	}
}

func (a *acquirerImpl) asyncTimeouter() {
	for {
		payments, err := a.s.List(PaymentState3dSecureRequired)
		if err != nil {
			log.Printf("[ERR] could not list payments: %v", err)
		}
		log.Println(payments)

		for _, payment := range payments {
			if payment.UpdatedAt.Add(time.Minute).Before(time.Now()) {
				_, err = a.CancelPayment(payment.Id, payment.Version)
				if err != nil {
					log.Printf("[ERR] failed to cancel payment %s: %s", payment.Id, err)
				}
			}
		}

		time.Sleep(10 * time.Second)
	}
}
