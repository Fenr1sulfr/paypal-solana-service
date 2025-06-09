import { Request, Response } from 'express';
import BigNumber from 'bignumber.js';
import { PublicKey, Keypair, Connection } from '@solana/web3.js';
import { encodeURL, findReference, validateTransfer } from '@solana/pay';
import { MongoClient } from 'mongodb';
const myWallet = 'DemoKMZWkk483hX4mUrcJoo3zVvsKhm8XXs28TuwZw9H'; // Replace with your wallet address (this is the destination where the payment will be sent)
const recipient = new PublicKey(myWallet);
const amount = new BigNumber(0.0001); // 0.0001 SOL
const label = 'QuickNode Guide Store';
const memo = 'QN Solana Pay Demo Public Memo';
const quicknodeEndpoint = 'https://api.devnet.solana.com'
const paymentRequests = new Map<string, { recipient: PublicKey; amount: BigNumber; memo: string }>();

async function generateUrl(
  recipient: PublicKey,
  amount: BigNumber,
  reference: PublicKey,
  label: string,
  message: string,
  memo: string,
) {
  const url: URL = encodeURL({
    recipient,
    amount,
    reference,
    label,
    message,
    memo,
  });
  return { url };
}

async function verifyTransaction(reference: PublicKey) {
  // 1 - Check that the payment request exists
  const paymentData = paymentRequests.get(reference.toBase58());
  if (!paymentData) {
    throw new Error('Payment request not found');
  }
  const { recipient, amount, memo } = paymentData;
  // 2 - Establish a Connection to the Solana Cluster
  
  const connection = new Connection(quicknodeEndpoint, 'confirmed');
  console.log('recipient', recipient.toBase58());
  console.log('amount', amount);
  console.log('reference', reference.toBase58());
  console.log('memo', memo);

  // 3 - Find the transaction reference
  
  const found = await findReference(connection, reference);
  console.log(found.signature)

  // 4 - Validate the transaction
  
  const response = await validateTransfer(
    connection,
    found.signature,
    {
      recipient,
      amount,
      splToken: undefined,
      reference,
      //memo
    },
    { commitment: 'confirmed' }
  );
  // 5 - Delete the payment request from local storage and return the response
  if (response) {
    paymentRequests.delete(reference.toBase58());
  }
  return response;
}

const mongoUri = process.env.MONGODB_URI || 'mongodb://localhost:27017';
const mongoDbName = process.env.MONGODB_DATABASE || 'payment_api';
let mongoClient: MongoClient | null = null;

async function getMongoClient() {
    if (!mongoClient) {
        mongoClient = new MongoClient(mongoUri);
        await mongoClient.connect();
    }
    return mongoClient;
}

export default async function SolanaPayHandler(req: Request, res: Response) {
    if (req.method === 'POST') {
        try {
            const reference = new Keypair().publicKey;
            const message = `QuickNode Demo - Order ID #0${Math.floor(Math.random() * 999999) + 1}`;
            const urlData = await generateUrl(
                recipient,
                amount,
                reference,
                label,
                message,
                memo
            );
            const ref = reference.toBase58();
            paymentRequests.set(ref, { recipient, amount, memo });
            const { url } = urlData;
            res.status(200).json({ url: url.toString(), ref });
        } catch (error) {
            console.error('Error:', error);
            res.status(500).json({ error: 'Internal Server Error' });
        }
    } 
    //...
}

export async function SolanaVerifyHandler(req: Request, res: Response) {
    if (req.method === 'GET') {
        const reference = req.query.reference;
        if (!reference) {
            res.status(400).json({ error: 'Missing reference query parameter' });
            return;
        }
        try {
            const referencePublicKey = new PublicKey(reference as string);
            const response = await verifyTransaction(referencePublicKey);
            if (response) {
                // Save verified transaction to MongoDB
                try {
                    const client = await getMongoClient();
                    const db = client.db(mongoDbName);
                    const collection = db.collection('sol_payments');
                    await collection.insertOne({
                        reference: referencePublicKey.toBase58(),
                        recipient: recipient.toBase58(),
                        amount: amount.toString(),
                        memo,
                        signature: response,
                        verifiedAt: new Date(),
                    });
                } catch (mongoError) {
                    console.error('MongoDB Error:', mongoError);
                }
                res.status(200).json({ status: 'verified' });
            } else {
                res.status(404).json({ status: 'not found' });
            }
        } catch (error) {
            console.error('Error:', error);
            res.status(500).json({ error: 'Internal Server Error' });
        }
    }
}
