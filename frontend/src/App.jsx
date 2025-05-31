import React, { useState } from 'react';
import {
  Container,
  Box,
  Typography,
  TextField,
  Button,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  CircularProgress,
} from '@mui/material';
import axios from 'axios';

function App() {
  const [quantity, setQuantity] = useState('');
  const [result, setResult] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const handleCalculate = async () => {
    if (!quantity || isNaN(quantity) || quantity <= 0) {
      setError('Please enter a valid quantity greater than 0');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const response = await axios.get(`/calculate?quantity=${quantity}`);
      setResult(response.data);
      setQuantity('');
    } catch (err) {
      setError(err.response?.data?.error || 'An error occurred while calculating pack allocation');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container maxWidth="md">
      <Box sx={{ my: 4 }}>
        <Typography variant="h3" component="h1" gutterBottom align="center">
          Smart Pack Allocation
        </Typography>
        
        <Box sx={{ 
          p: 3, 
          mb: 3, 
          backgroundColor: 'background.paper'
        }}>
          <Box 
            component="form" 
            onSubmit={(e) => {
              e.preventDefault();
              handleCalculate();
            }}
            sx={{ display: 'flex', gap: 2, alignItems: 'center' }}
          >
            <TextField
              label="Order Quantity"
              type="number"
              value={quantity}
              onChange={(e) => setQuantity(e.target.value)}
              fullWidth
              error={!!error}
              helperText={error}
            />
            <Button
              type="submit"
              variant="contained"
              disabled={loading}
              sx={{ minWidth: 120 }}
            >
              {loading ? <CircularProgress size={24} /> : 'Calculate'}
            </Button>
          </Box>
        </Box>

        {result && (
          <Box sx={{ 
            p: 3,
            backgroundColor: 'background.paper'
          }}>
            <Typography variant="h6" gutterBottom>
              Pack Allocation Result
            </Typography>
            <TableContainer>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>Pack Size</TableCell>
                    <TableCell align="right">Quantity</TableCell>
                    <TableCell align="right">Total Items</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {Object.entries(result.packs).map(([size, count]) => (
                    <TableRow key={size}>
                      <TableCell>{size}</TableCell>
                      <TableCell align="right">{count}</TableCell>
                      <TableCell align="right">{size * count}</TableCell>
                    </TableRow>
                  ))}
                  <TableRow>
                    <TableCell colSpan={2} align="right">
                      <strong>Total Items:</strong>
                    </TableCell>
                    <TableCell align="right">
                      <strong>{result.total}</strong>
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </TableContainer>
          </Box>
        )}
      </Box>
    </Container>
  );
}

export default App; 