export interface PaymentData {
  paymentId: string;
  paymentOwner: string;
  paymentSender: string;
  paymentTimestamp: number;
  paymentAmount: number;
  paymentCurrency: number[];
  paymentUrl: string;
}
