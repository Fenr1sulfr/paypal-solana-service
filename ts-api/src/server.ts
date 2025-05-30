import express from "express";
import cors from "cors";
import { handleInitialize } from "./api/initialize";
import { handleAddPayment } from "./api/addPayment";

const app = express();
app.use(cors());
app.use(express.json());

app.post("/initialize", handleInitialize);
app.post("/add-payment", handleAddPayment);

const PORT = process.env.PORT || 8081;
app.listen(PORT, () => console.log(`Server is running on http://localhost:${PORT}`));
