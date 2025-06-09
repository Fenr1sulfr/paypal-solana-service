import {
  createNft,
  fetchDigitalAsset,
  mplTokenMetadata,
} from "@metaplex-foundation/mpl-token-metadata";
import {
  airdropIfRequired,
  getExplorerLink,
  getKeypairFromFile,
} from "@solana-developers/helpers";
import { createUmi } from "@metaplex-foundation/umi-bundle-defaults";
import { Connection, Keypair, LAMPORTS_PER_SOL, clusterApiUrl } from "@solana/web3.js";
import {
  generateSigner,
  keypairIdentity,
  percentAmount,
  publicKey,
} from "@metaplex-foundation/umi";

export async function createNftForRecipient(recipientAddress: string): Promise<string> {
    const connection = new Connection(clusterApiUrl("devnet"), "confirmed");
    const user = Keypair.fromSecretKey(Uint8Array.from());

    await airdropIfRequired(
        connection,
        user.publicKey,
        1 * LAMPORTS_PER_SOL,
        0.5 * LAMPORTS_PER_SOL
    );

    const umi = createUmi(connection.rpcEndpoint);
    umi.use(mplTokenMetadata());

    const umiUser = umi.eddsa.createKeypairFromSecretKey(user.secretKey);
    umi.use(keypairIdentity(umiUser));

    const collectionAddress = publicKey(
        "A5zE9JrWzG5hUjiHBJWbtkRzynxkj86sEUYR4b9UxjAW"
    );

    const mint = generateSigner(umi);
    const recipient = publicKey(recipientAddress);

    const transaction = await createNft(umi, {
        mint,
        name: "Agrichain NFT",
        uri: "https://raw.githubusercontent.com/Fenr1sulfr/paypal-solana-service/main/metadata.json",
        sellerFeeBasisPoints: percentAmount(0),
        collection: {
            key: collectionAddress,
            verified: false,
        },
        tokenOwner: recipient,
    });

    await transaction.sendAndConfirm(umi);

    // Retry logic to fetch the created NFT
    async function fetchWithRetry(umi: any, publicKey: any, maxRetries = 10, delay = 2000) {
        for (let i = 0; i < maxRetries; i++) {
            try {
                const asset = await fetchDigitalAsset(umi, publicKey);
                return asset;
            } catch (error: any) {
                if (error.message?.includes('AccountNotFoundError') || error.message?.includes('was not found')) {
                    if (i < maxRetries - 1) {
                        await new Promise(resolve => setTimeout(resolve, delay));
                        delay = Math.min(delay * 1.2, 10000);
                    }
                } else {
                    throw error;
                }
            }
        }
        throw new Error(`Failed to fetch digital asset after ${maxRetries} attempts`);
    }

    const createdNft = await fetchWithRetry(umi, mint.publicKey);

    return getExplorerLink(
        "address",
        createdNft.mint.publicKey,
        "devnet"
    );
}
