import { Connection, Keypair, LAMPORTS_PER_SOL, PublicKey, sendAndConfirmTransaction, SystemProgram, Transaction, TransactionInstruction } from '@solana/web3.js';
import { encodeURL, validateTransfer, parseURL, TransferRequestURL, findReference } from '@solana/pay';
import BigNumber from 'bignumber.js';


const secret = [32,227,51,228,176,83,158,245,151,9,58,107,172,157,183,159,73,253,128,254,37,179,46,153,42,80,206,35,175,209,254,255,112,50,143,212,143,8,103,173,88,102,201,136,0,69,4,208,221,84,144,42,235,97,66,16,204,95,28,84,14,94,243,100]; // Replace with your secret key
const payer = Keypair.fromSecretKey(new Uint8Array(secret));

const myWallet = 'DemoKMZWkk483hX4mUrcJoo3zVvsKhm8XXs28TuwZw9H'; // Replace with your wallet address (this is the destination where the payment will be sent)
const recipient = new PublicKey(myWallet);
const quickNodeEndpoint = 'https://api.devnet.solana.com'; // Replace with your QuickNode endpoint
const connection = new Connection(quickNodeEndpoint, 'confirmed');
const amount = new BigNumber(0.1); // 0.1 SOL
const reference = new Keypair().publicKey;
const label = 'Agrichain receipt';
const message = `Agrichain Demo - Order ID #0${Math.floor(Math.random() * 999999) + 1}`;
const memo = 'AGRICHAIN Solana Pay Demo Public Memo';

export async function generateUrl(
    recipient: PublicKey,
    amount: BigNumber,
    reference: PublicKey,
    label: string,
    message: string,
    memo: string
) {
    console.log('1. Create a payment request link');
    const url: URL = encodeURL({ recipient, amount, reference, label, message, memo });
    console.log('Payment request link:', url);
    return url;
}

// Создать и вернуть сериализованную (base64) транзакцию для фронта
export async function createUnsignedTransaction(
    recipient: PublicKey,
    amount: BigNumber,
    reference: PublicKey,
    memo: string,
    payer: PublicKey // публичный ключ плательщика (отправителя)
): Promise<string> {
    const tx = new Transaction();

    // Добавить memo, если есть
    if (memo != null) {
        tx.add(
            new TransactionInstruction({
                programId: new PublicKey('MemoSq4gqABAXKb96qnH8TysNcWxMyWCqXgDLGmfcHr'),
                keys: [],
                data: Buffer.from(memo, 'utf8'),
            })
        );
    }

    // Инструкция перевода SOL
    const ix = SystemProgram.transfer({
        fromPubkey: payer,
        toPubkey: recipient,
        lamports: amount.multipliedBy(LAMPORTS_PER_SOL).integerValue(BigNumber.ROUND_FLOOR).toNumber()
    });

    // Добавить reference
    if (reference) {
        const ref = Array.isArray(reference) ? reference : [reference];
        for (const pubkey of ref) {
            ix.keys.push({ pubkey, isWritable: false, isSigner: false });
        }
    }
    tx.add(ix);

    // Получить последний блокхеш для валидности транзакции
    const { blockhash } = await connection.getLatestBlockhash('finalized');
    tx.recentBlockhash = blockhash;
    tx.feePayer = payer;

    // Сериализовать транзакцию (без подписей)
    const serialized = tx.serialize({ requireAllSignatures: false, verifySignatures: false });
    return serialized.toString('base64');
}

export async function processPayment(url: URL, payer: Keypair) {
    // Parse the payment request link
    console.log('2. Parse the payment request link');
    const { recipient, amount, reference, label, message, memo } = parseURL(url) as TransferRequestURL;
    if (!recipient || !amount || !reference) throw new Error('Invalid payment request link');
    
    console.log('3. Assemble the transaction');
    const tx = new Transaction();

    // Append the memo instruction if a memo is provided
    if (memo != null) {
        tx.add(
            new TransactionInstruction({
                programId: new PublicKey('MemoSq4gqABAXKb96qnH8TysNcWxMyWCqXgDLGmfcHr'),
                keys: [],
                data: Buffer.from(memo, 'utf8'),
            })
        );
    }
    // Create a transfer instruction
    const ix = SystemProgram.transfer({
        fromPubkey: payer.publicKey,
        toPubkey: recipient,
        lamports: amount.multipliedBy(LAMPORTS_PER_SOL).integerValue(BigNumber.ROUND_FLOOR).toNumber()
    });
    // Add the reference key to the instruction, if provided
    if (reference) {
        const ref = Array.isArray(reference) ? reference : [reference];
        for (const pubkey of ref) {
            ix.keys.push({ pubkey, isWritable: false, isSigner: false });
        }
    }
    // Add the transfer instruction to the transaction
    tx.add(ix);

    // Send the transaction to the Solana network and confirm it has been processed
    console.log('4. 🚀 Send and Confirm Transaction');
    const txId = await sendAndConfirmTransaction(connection, tx, [payer]);
    console.log(`      Tx: https://explorer.solana.com/tx/${txId}?cluster=devnet`);
}

export async function verifyTx(
    recipient: PublicKey,
    amount: BigNumber,
    reference: PublicKey,
    memo: string
) {
    console.log(`5. Verifying the payment`);
    // Merchant app locates the transaction signature from the unique reference address it provided in the transfer link
    const found = await findReference(connection, reference);

    // Merchant app should always validate that the transaction transferred the expected amount to the recipient
    const response = await validateTransfer(
        connection,
        found.signature,
        {
            recipient,
            amount,
            splToken: undefined,
            reference,
            memo
        },
        { commitment: 'confirmed' }
    );
    return response;
}